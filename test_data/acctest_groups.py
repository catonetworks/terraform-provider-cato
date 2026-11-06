#!/usr/bin/env python3
"""Balance acceptance suites and select one account without exposing the pool."""

import argparse
import json
import math
import os
from pathlib import Path
import random
import re
import sys

ROOT = Path(__file__).resolve().parent.parent
TIMINGS = ROOT / 'test_data/acctest_timings.json'
GROUP_COUNT = 10
FIXTURE_TYPES = ('custom_apps', 'custom_categories', 'device_postures',
                 'floating_ranges', 'system_groups', 'user_groups')
# Only fixture settings may be supplied per account; never override credentials,
# the endpoint, runner paths, or arbitrary workflow environment variables.
TEST_ENV_NAMES = frozenset((
    'CATO_TEST_SITE_ID', 'TFACC_ENABLE_DNS_DHCP_CRUD',
    'TFACC_DNS_DHCP_ALLOWED_ACCOUNT_ID', 'TFACC_DNS_DHCP_SERVERS',
    'TFACC_DHCP_OPTION_TAG', 'TFACC_SDP_STATIC_IP', 'TFACC_SDP_STATIC_IP_ANCHOR',
    'TFACC_ALLOW_UNREGISTERED_SECONDARY_SITE', 'ACC_TEST_BW_LICENSE',
    'TFACC_FQDN_CONTAINER_URL', 'TFACC_IP_RANGE_CONTAINER_URL',
    'TFACC_CONTAINER_NOTIFICATION_MAILING_LIST_ID', 'TFACC_CONTAINER_URL_ALLOW_NOTIFICATIONS',
))


def load_timings():
    return json.loads(TIMINGS.read_text())


def workloads(timings):
    units = []
    for suite, timing in sorted(timings['packages'].items()):
        if not re.fullmatch(r'[a-z0-9_]+', suite):
            raise ValueError('Invalid suite name in timing inventory')
        for test, seconds in timing.get('tests', {None: timing['seconds']}).items():
            if test is not None and not re.fullmatch(r'Test\w+', test):
                raise ValueError('Invalid test name in timing inventory')
            if not isinstance(seconds, (int, float)) or not math.isfinite(seconds) or seconds < 0:
                raise ValueError('Invalid test duration')
            units.append({'suite': suite, 'test': test, 'seconds': seconds})
    return units


def plan():
    timings = load_timings()
    overhead = timings['invocation_overhead_seconds']
    result = [{'id': i, 'seconds': 0, 'units': []} for i in range(GROUP_COUNT)]
    for unit in sorted(workloads(timings), key=lambda u: (-u['seconds'], u['suite'], u['test'] or '')):
        def cost(group):
            extra = 0 if any(u['suite'] == unit['suite'] for u in group['units']) else overhead
            return group['seconds'] + unit['seconds'] + extra, group['id']
        group = min(result, key=cost)
        group['seconds'] = cost(group)[0]
        group['units'].append(unit)
    return result


def validate_inventory():
    packages = load_timings()['packages']
    discovered = {p.parent.name for p in (ROOT / 'internal/acctests').rglob('*_test.go')
                  if p.parent.name != 'acc'}
    if set(packages) != discovered:
        raise ValueError(f'Update acctest_timings.json: missing={sorted(discovered - set(packages))}, '
                         f'removed={sorted(set(packages) - discovered)}')
    for suite, timing in packages.items():
        if 'tests' not in timing:
            continue
        declared = set()
        for source in (ROOT / 'internal/acctests' / suite).glob('*_test.go'):
            declared.update(re.findall(r'^func (Test\w+)\(t \*testing\.T\)', source.read_text(), re.MULTILINE))
        if declared != set(timing['tests']):
            raise ValueError(f'Update individual test timings for {suite}')


def matrix(rng=None):
    slots = list(range(GROUP_COUNT))
    (rng or random.SystemRandom()).shuffle(slots)
    return {'include': [{'group': group, 'account_slot': slot} for group, slot in enumerate(slots)]}


def test_filter(group, suite):
    selected = [u for u in plan()[group]['units'] if u['suite'] == suite]
    if not selected:
        raise ValueError('Suite does not belong to this group')
    if any(u['test'] is None for u in selected):
        return '.'
    return '^(' + '|'.join(sorted(u['test'] for u in selected)) + ')$'


def validate_accounts(accounts):
    if not isinstance(accounts, list) or len(accounts) != GROUP_COUNT:
        raise ValueError('CATO_ACCTEST_ACCOUNTS must contain exactly ten accounts')
    ids, keys = set(), set()
    for account in accounts:
        if not isinstance(account, dict):
            raise ValueError('Invalid account record')
        account_id, key = account.get('id'), account.get('api_key')
        if not isinstance(account_id, str) or not account_id.isdecimal():
            raise ValueError('Account IDs must be numeric strings')
        if not isinstance(key, str) or not key or any(c in key for c in '\r\n'):
            raise ValueError('Account API keys must be nonempty single-line strings')
        if account_id in ids or key in keys:
            raise ValueError('Accounts and API keys must be unique')
        if not isinstance(account.get('test_vars', {}), dict):
            raise ValueError('Per-account test_vars must be an object')
        test_env = account.get('test_env', {})
        if not isinstance(test_env, dict):
            raise ValueError('Per-account test_env must be an object')
        for name, value in test_env.items():
            if name not in TEST_ENV_NAMES or not isinstance(value, str) or any(c in value for c in '\r\n'):
                raise ValueError('Invalid per-account fixture environment setting')
        if test_env.get('TFACC_DNS_DHCP_ALLOWED_ACCOUNT_ID', account_id) != account_id:
            raise ValueError('DNS/DHCP fixture authorization must match the selected account')
        ids.add(account_id)
        keys.add(key)


def write_account_environment(accounts, slot, path):
    validate_accounts(accounts)
    if slot not in range(GROUP_COUNT):
        raise ValueError('Account slot must be between 0 and 9')
    account = accounts[slot]
    # JSON encoding keeps fixture values on one line in the GitHub environment file.
    with Path(path).open('a') as output:
        output.write(f'CATO_ACCOUNT_ID={account["id"]}\nCATO_TOKEN={account["api_key"]}\n')
        output.write('TFACC_TEST_VARS=' + json.dumps(account.get('test_vars', {}), separators=(',', ':')) + '\n')
        for name, value in sorted(account.get('test_env', {}).items()):
            output.write(f'{name}={value}\n')


def validate_fixture_references(accounts):
    for account in accounts:
        for fixture_type in FIXTURE_TYPES:
            refs = account.get('test_vars', {}).get(fixture_type)
            if not isinstance(refs, list) or len(refs) < 3 or any(
                    not isinstance(ref, dict) or not isinstance(ref.get('id'), str) or not ref['id']
                    or not isinstance(ref.get('name'), str) or not ref['name'] for ref in refs):
                raise ValueError(f'Every CI account needs at least three {fixture_type} fixture references')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['validate', 'matrix', 'suites', 'filter', 'account', 'summary'])
    parser.add_argument('--group', type=int, choices=range(GROUP_COUNT))
    parser.add_argument('--slot', type=int, choices=range(GROUP_COUNT))
    parser.add_argument('--suite')
    args = parser.parse_args()
    validate_inventory()
    if args.command == 'validate':
        return
    if args.command == 'matrix':
        print(json.dumps(matrix(), separators=(',', ':')))
    elif args.command == 'summary':
        for group in plan():
            suites = sorted({u['suite'] for u in group['units']})
            print(f'Group {group["id"]}: {group["seconds"]:.2f}s; ' + ', '.join(suites))
    elif args.command in ['suites', 'filter']:
        if args.group is None:
            parser.error('--group is required')
        if args.command == 'suites':
            print('\n'.join(sorted({u['suite'] for u in plan()[args.group]['units']})))
        else:
            print(test_filter(args.group, args.suite))
    else:
        if args.slot is None:
            parser.error('--slot is required')
        try:
            accounts = json.loads(os.environ['CATO_ACCTEST_ACCOUNTS'])
        except (KeyError, json.JSONDecodeError):
            raise ValueError('CATO_ACCTEST_ACCOUNTS is missing or invalid JSON') from None
        # Mask individual keys as well as the complete JSON secret before any tests run.
        validate_accounts(accounts)
        validate_fixture_references(accounts)
        for account in accounts:
            key = account['api_key'].replace('%', '%25')
            print(f'::add-mask::{key}')
        write_account_environment(accounts, args.slot, os.environ['GITHUB_ENV'])
        print(f'Acceptance account slot {args.slot}, account ID {accounts[args.slot]["id"]}')


if __name__ == '__main__':
    try:
        main()
    except ValueError as error:
        sys.exit(str(error))
