import copy
import json
import os
import pathlib
import random
import subprocess
import tempfile
import unittest

import acctest_groups as groups


class AcceptanceGroupsTest(unittest.TestCase):
    def test_complete_balanced_partition(self):
        plan = groups.plan()
        self.assertEqual(len(plan), 10)
        # Keep the balance check proportional as measured durations change.
        average_seconds = sum(g['seconds'] for g in plan) / len(plan)
        self.assertLess(max(g['seconds'] for g in plan) - min(g['seconds'] for g in plan),
                        average_seconds * 0.05)
        expected = groups.workloads(groups.load_timings())
        actual = [unit for group in plan for unit in group['units']]
        self.assertCountEqual([(u['suite'], u.get('test')) for u in actual],
                              [(u['suite'], u.get('test')) for u in expected])
        groups.validate_inventory()

    def test_random_assignment_is_a_permutation(self):
        a = groups.matrix(random.Random(1))
        b = groups.matrix(random.Random(2))
        self.assertEqual(sorted(x['account_slot'] for x in a['include']), list(range(10)))
        self.assertEqual([x['group'] for x in a['include']], list(range(10)))
        self.assertNotEqual(a, b)

    def test_split_suite_filters_are_anchored_and_exhaustive(self):
        for suite, timing in groups.load_timings()['packages'].items():
            if 'tests' not in timing:
                continue
            with self.subTest(suite=suite):
                tests = []
                for group in groups.plan():
                    units = [u for u in group['units'] if u['suite'] == suite]
                    if units:
                        pattern = groups.test_filter(group['id'], suite)
                        self.assertTrue(pattern.startswith('^('))
                        self.assertTrue(pattern.endswith(')$'))
                        tests.extend(u['test'] for u in units)
                self.assertCountEqual(tests, timing['tests'])

    def test_account_pool_rejects_duplicates_and_partial_pools(self):
        pool = [{'id': str(i), 'api_key': f'inert-test-key-{i}'} for i in range(10)]
        groups.validate_accounts(pool)
        with self.assertRaises(ValueError):
            groups.validate_accounts(pool[:9])
        duplicate = copy.deepcopy(pool)
        duplicate[1]['id'] = duplicate[0]['id']
        with self.assertRaises(ValueError):
            groups.validate_accounts(duplicate)
        duplicate = copy.deepcopy(pool)
        duplicate[1]['api_key'] = duplicate[0]['api_key']
        with self.assertRaises(ValueError):
            groups.validate_accounts(duplicate)

    def test_account_environment_has_only_selected_credentials_and_vars(self):
        pool = [{'id': str(i), 'api_key': f'inert-test-key-{i}'} for i in range(10)]
        pool[3]['test_vars'] = {'user_groups': [{'id': '31', 'name': 'fixture'}]}
        with tempfile.TemporaryDirectory() as directory:
            env_file = pathlib.Path(directory) / 'env'
            groups.write_account_environment(pool, 3, env_file)
            content = env_file.read_text()
            self.assertIn('CATO_ACCOUNT_ID=3\n', content)
            self.assertIn('CATO_TOKEN=inert-test-key-3\n', content)
            self.assertNotIn('inert-test-key-2', content)
            self.assertIn('"user_groups"', content)
            groups.write_account_environment(pool, 2, env_file)
            self.assertTrue(env_file.read_text().endswith('TFACC_TEST_VARS={}\n'))

    def test_account_environment_includes_only_safe_fixture_settings(self):
        pool = [{'id': str(i), 'api_key': f'inert-test-key-{i}'} for i in range(10)]
        pool[3]['test_env'] = {'TFACC_ENABLE_DNS_DHCP_CRUD': 'true',
                               'TFACC_DNS_DHCP_ALLOWED_ACCOUNT_ID': '3',
                               'CATO_TEST_SITE_ID': '42'}
        with tempfile.TemporaryDirectory() as directory:
            output = pathlib.Path(directory) / 'environment'
            groups.write_account_environment(pool, 3, output)
            self.assertIn('TFACC_ENABLE_DNS_DHCP_CRUD=true\n', output.read_text())
            self.assertIn('CATO_TEST_SITE_ID=42\n', output.read_text())
        for settings in [{'CATO_TOKEN': 'override'}, {'PATH': '/tmp'},
                         {'TFACC_ENABLE_DNS_DHCP_CRUD': 'true\nCATO_TOKEN=override'},
                         {'TFACC_DNS_DHCP_ALLOWED_ACCOUNT_ID': '4'}]:
            with self.subTest(settings=settings):
                pool[3]['test_env'] = settings
                with self.assertRaises(ValueError):
                    groups.validate_accounts(pool)

    def test_ci_rejects_missing_or_incomplete_account_fixtures(self):
        pool = [{'id': str(i), 'api_key': f'inert-test-key-{i}'} for i in range(10)]
        with self.assertRaisesRegex(ValueError, 'fixture references'):
            groups.validate_fixture_references(pool)
        for account in pool:
            account['test_vars'] = {kind: [{'id': str(n), 'name': f'fixture-{n}'} for n in range(3)]
                                    for kind in groups.FIXTURE_TYPES}
        groups.validate_fixture_references(pool)
        pool[9]['test_vars']['floating_ranges'][0]['id'] = ''
        with self.assertRaisesRegex(ValueError, 'floating_ranges'):
            groups.validate_fixture_references(pool)

    def test_make_exports_acceptance_switches(self):
        names = ('TFACC_ENABLE_ACCOUNT_CRUD', 'TFACC_ACCOUNT_CRUD_ALLOWED',
                 'TFACC_ENABLE_BGP_PEER_CRUD', 'TFACC_ENABLE_RULES_INDEX_CRUD')
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            (root / 'test_data').mkdir()
            runner = root / 'test_data/flaky_acctest.sh'
            runner.write_text("""#!/usr/bin/env python3
import json, os
names = ('TFACC_ENABLE_ACCOUNT_CRUD', 'TFACC_ACCOUNT_CRUD_ALLOWED',
         'TFACC_ENABLE_BGP_PEER_CRUD', 'TFACC_ENABLE_RULES_INDEX_CRUD')
with open(os.environ['TEST_SWITCH_OUTPUT'], 'w') as output:
    json.dump({name: os.environ.get(name) for name in names}, output)
""")
            runner.chmod(0o700)
            output = root / 'switches.json'
            env = {key: value for key, value in os.environ.items() if key not in names}
            env['TEST_SWITCH_OUTPUT'] = str(output)
            command = ['make', '-f', str(groups.ROOT / 'GNUmakefile'), 'acctest-flaky']
            subprocess.run(command, cwd=root, env=env, capture_output=True, check=True)
            self.assertEqual(json.loads(output.read_text()), dict(zip(names, ('false', 'false', 'true', 'true'))))
            env.update(dict.fromkeys(names, 'true'))
            subprocess.run(command, cwd=root, env=env, capture_output=True, check=True)
            self.assertEqual(json.loads(output.read_text()), dict.fromkeys(names, 'true'))

    def test_runner_keeps_groups_and_single_account_runs_serial(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            binaries = root / 'bin'
            binaries.mkdir()
            calls = root / 'calls.jsonl'
            go = binaries / 'go'
            go.write_text('''#!/usr/bin/env python3
import json, os, sys
with open(os.environ['TEST_GO_CALLS'], 'a') as output:
    output.write(json.dumps(sys.argv[1:]) + '\\n')
if sys.argv[1:3] == ['tool', 'tparse']:
    sys.stdin.read()
''')
            go.chmod(0o700)
            for suite in groups.load_timings()['packages']:
                package = root / 'internal/acctests' / suite
                package.mkdir(parents=True)
                (package / 'fixture_test.go').touch()
            env = dict(os.environ, PATH=str(binaries) + os.pathsep + os.environ['PATH'],
                       TEST_GO_CALLS=str(calls))
            runner = str(groups.ROOT / 'test_data/flaky_acctest.sh')
            for selected in [[], ['--group', '1'], ['admin']]:
                with self.subTest(selected=selected):
                    calls.write_text('')
                    subprocess.run(['bash', runner] + selected, cwd=root, env=env,
                                   capture_output=True, text=True, check=True)
                    commands = [json.loads(line) for line in calls.read_text().splitlines()]
                    tests = [c for c in commands if c[0] == 'test' and '-parallel=1' in c]
                    expected = (sorted(groups.load_timings()['packages']) if not selected else
                                sorted({u['suite'] for u in groups.plan()[1]['units']})
                                if selected[0] == '--group' else ['admin'])
                    self.assertEqual([next(a.rsplit('/', 1)[1] for a in c
                                           if a.startswith('./internal/acctests/')) for c in tests], expected)
                    for command in tests:
                        self.assertIn('-p=1', command)
                        suite = next(a.rsplit('/', 1)[1] for a in command
                                     if a.startswith('./internal/acctests/'))
                        pattern = command[command.index('-run') + 1]
                        self.assertEqual(pattern, groups.test_filter(1, suite) if selected[:1] == ['--group'] else '.')


if __name__ == '__main__':
    unittest.main()
