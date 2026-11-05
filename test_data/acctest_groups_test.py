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
        self.assertLess(max(g['seconds'] for g in plan) - min(g['seconds'] for g in plan), 5)
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
        tests = []
        for group in groups.plan():
            units = [u for u in group['units'] if u['suite'] == 'socket_site']
            if units:
                pattern = groups.test_filter(group['id'], 'socket_site')
                self.assertTrue(pattern.startswith('^('))
                self.assertTrue(pattern.endswith(')$'))
                tests.extend(u['test'] for u in units)
        self.assertCountEqual(tests, groups.load_timings()['packages']['socket_site']['tests'])

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
