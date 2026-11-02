#!/usr/bin/env python3
"""Read baseline and restore only DNS/DHCP settings in an explicitly selected dev account."""
import argparse
import json
import os
from pathlib import Path
import urllib.parse
import urllib.request


def configuration():
    settings = {name: os.environ.get(name, "") for name in
                ("CATO_BASEURL", "CATO_ACCOUNT_ID", "CATO_TEST_SITE_ID", "CATO_TOKEN")}
    missing = [name for name, value in settings.items() if not value]
    if missing:
        raise ValueError("Missing environment variables: " + ", ".join(missing))
    url = urllib.parse.urlparse(settings["CATO_BASEURL"])
    if (url.scheme != "https" or url.hostname != "system.cc.dev-us1-m1.catonet.works"
            or url.path != "/api/v1/graphql" or url.username or url.query or url.fragment):
        raise ValueError("This fixture is restricted to the dev-us1-m1 HTTPS GraphQL endpoint")
    for source, destination in (("CATO_BASEURL", "TF_VAR_baseurl"),
                                ("CATO_ACCOUNT_ID", "TF_VAR_account_id"),
                                ("CATO_TEST_SITE_ID", "TF_VAR_site_id")):
        if os.environ.get(destination, settings[source]) != settings[source]:
            raise ValueError(source + " must match " + destination)
    return settings


def query(document, *, variables=None):
    config = configuration()
    request = urllib.request.Request(config["CATO_BASEURL"], method="POST",
        data=json.dumps({"query": document, "variables": variables or {}}).encode(),
        headers={"Content-Type": "application/json", "x-api-key": config["CATO_TOKEN"],
                 "x-account-id": config["CATO_ACCOUNT_ID"]})
    with urllib.request.urlopen(request, timeout=30) as response:
        result = json.load(response)
    if result.get("errors"):
        raise ValueError("GraphQL preflight failed: " + "; ".join(
            str(error.get("message", "unknown error")) for error in result["errors"]))
    if result.get("data") is None:
        raise ValueError("API returned no data")
    return result["data"]


READ_SETTINGS = """query ReadDNSDHCP($account: ID!, $site: SiteRefInput!) {
 networkConfig(accountId: $account) {
  dns { settings { primaryServer secondaryServer suffix acceptDnsRequestsOnLanInterfaceIp }
   siteSettings(site: $site) { siteId primaryServer secondaryServer suffix } }
  dhcp { settings { leaseTime relay { enabled timeout group { id name } } } }
 }
}"""


def read_settings():
    config = configuration()
    result = query(READ_SETTINGS, variables={"account": config["CATO_ACCOUNT_ID"],
                   "site": {"by": "ID", "input": config["CATO_TEST_SITE_ID"]}})
    data = result.get("networkConfig")
    if not data or not data.get("dns", {}).get("settings") or not data.get("dhcp", {}).get("settings"):
        raise ValueError("Missing DNS/DHCP settings payload")
    return data


def collection(domain, name, fields):
    config = configuration()
    items, offset = [], 0
    while True:
        document = ("query BaselineList($account: ID!, $offset: Int!) { networkConfig(accountId: $account) { "
            + domain + " { " + name + "(input: {paging: {from: $offset, limit: 100}}) { items { "
            + fields + " } paging { total } } } } }")
        data = query(document, variables={"account": config["CATO_ACCOUNT_ID"], "offset": offset})
        payload = data["networkConfig"][domain][name]
        page = payload["items"]
        items.extend(page)
        if len(items) >= payload["paging"]["total"]:
            return items
        if not page:
            raise ValueError("Incomplete collection pagination")
        offset += len(page)


def snapshot():
    config = configuration()
    # Entity lookup verifies the site under this account even when DNS override is absent.
    sites = query("""query SiteOwnership($account: ID!, $ids: [ID!]) {
      entityLookup(accountID: $account, type: site, entityIDs: $ids) { items { entity { id } } }
    }""", variables={"account": config["CATO_ACCOUNT_ID"], "ids": [config["CATO_TEST_SITE_ID"]]})
    if not any(item["entity"]["id"] == config["CATO_TEST_SITE_ID"] for item in sites["entityLookup"]["items"]):
        raise ValueError("Site does not belong to selected account")
    data = read_settings()
    data["collections"] = {
        "forwardingRule": collection("dns", "forwardingRuleList", "id domain server"),
        "option": collection("dhcp", "optionList", "id tag type value description"),
        "relayGroup": collection("dhcp", "relayGroupList", "id name server"),
    }
    data["target"] = {"account": config["CATO_ACCOUNT_ID"], "site": config["CATO_TEST_SITE_ID"],
                      "endpoint": config["CATO_BASEURL"]}
    return data


def restore_account(baseline):
    settings = baseline["dhcp"]["settings"]
    relay = {"enabled": settings["relay"]["enabled"]}
    if settings["relay"]["timeout"] is not None:
        relay["timeout"] = settings["relay"]["timeout"]
    if settings["relay"]["group"] is not None:
        relay["group"] = {"by": "ID", "input": settings["relay"]["group"]["id"]}
    query("""mutation RestoreAccountDNSDHCP($account: ID!, $dns: NetworkConfigDnsUpdateSettingsInput!,
         $dhcp: NetworkConfigDhcpUpdateSettingsInput!) { networkConfig(accountId: $account) {
       dns { updateSettings(input: $dns) { settings { primaryServer secondaryServer suffix } } }
       dhcp { updateSettings(input: $dhcp) { settings { leaseTime } } }
    } }""", variables={"account": os.environ.get("CATO_ACCOUNT_ID"), "dns": baseline["dns"]["settings"],
                       "dhcp": {"leaseTime": settings["leaseTime"], "relay": relay}})


def restore_site(baseline):
    target = baseline["target"]
    site = {"by": "ID", "input": target["site"]}
    original = baseline["dns"]["siteSettings"]
    if original is None:
        query("""mutation RestoreSite($account: ID!, $input: NetworkConfigDnsRemoveSiteSettingsInput!) {
          networkConfig(accountId: $account) { dns { removeSiteSettings(input: $input) { siteSettings { siteId } } } }
        }""", variables={"account": target["account"], "input": {"site": [site]}})
    else:
        settings = {key: original[key] for key in ("primaryServer", "secondaryServer", "suffix")}
        settings["site"] = site
        query("""mutation RestoreSite($account: ID!, $input: NetworkConfigDnsUpdateSiteSettingsInput!) {
          networkConfig(accountId: $account) { dns { updateSiteSettings(input: $input) { siteSettings { siteId } } } }
        }""", variables={"account": target["account"], "input": {"siteSettings": [settings]}})


def load_baseline(filename):
    data = json.loads(filename.read_text())
    config = configuration()
    expected = {"account": config["CATO_ACCOUNT_ID"], "site": config["CATO_TEST_SITE_ID"],
                "endpoint": config["CATO_BASEURL"]}
    if data.get("target") != expected:
        raise ValueError("Baseline target does not match current environment")
    return data


def assert_restored(baseline):
    actual = snapshot()
    for domain in ("dns", "dhcp"):
        if baseline[domain] != actual[domain]:
            raise ValueError(domain.upper() + " settings differ from baseline")
    for name, expected in baseline["collections"].items():
        if sorted(expected, key=lambda item: item["id"]) != sorted(actual["collections"][name], key=lambda item: item["id"]):
            raise ValueError(name + " collection differs from baseline")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("capture", "restore-account", "restore-site", "verify"))
    parser.add_argument("--baseline", type=Path, default=Path("baseline.json"))
    args = parser.parse_args()
    configuration()
    if args.command == "capture":
        data = snapshot()
        # Never overwrite original baseline. Restrict local configuration file permissions.
        fd = os.open(args.baseline, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
        with os.fdopen(fd, "w") as output:
            json.dump(data, output, indent=2)
        print("Baseline captured; read-only preflight passed")
    else:
        baseline = load_baseline(args.baseline)
        if args.command == "restore-account":
            restore_account(baseline)
        elif args.command == "restore-site":
            restore_site(baseline)
        else:
            assert_restored(baseline)
        print(args.command + " completed")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, KeyError) as error:
        raise SystemExit(str(error)) from None
