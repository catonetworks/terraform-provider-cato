import unittest
from unittest.mock import patch
import preflight

class PreflightTests(unittest.TestCase):
    def test_settings_restore_preserves_dns_clears_and_false(self):
        baseline = {"dns": {"settings": {"primaryServer": None, "secondaryServer": None,
                    "suffix": [], "acceptDnsRequestsOnLanInterfaceIp": False}},
                    "dhcp": {"settings": {"leaseTime": 3600, "relay": {"enabled": False,
                    "timeout": 0, "group": None}}}}
        with patch.object(preflight, "query", return_value={}) as call:
            preflight.restore_account(baseline)
        variables = call.call_args.kwargs["variables"]
        self.assertIsNone(variables["dns"]["primaryServer"])
        self.assertEqual([], variables["dns"]["suffix"])
        self.assertIs(False, variables["dns"]["acceptDnsRequestsOnLanInterfaceIp"])
        self.assertEqual(0, variables["dhcp"]["relay"]["timeout"])
        self.assertNotIn("group", variables["dhcp"]["relay"])

    def test_rejects_non_dev_endpoint(self):
        with patch.dict("os.environ", {"CATO_BASEURL": "https://api.catonetworks.com/api/v1/graphql",
                        "CATO_ACCOUNT_ID": "test", "CATO_TOKEN": "test-key", "CATO_TEST_SITE_ID": "site"}):
            with self.assertRaises(ValueError):
                preflight.configuration()

if __name__ == "__main__":
    unittest.main()
