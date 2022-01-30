import os
import sys
import subprocess

# TODO: amend to be less OS dependent
EXEC_PATH = os.path.join(os.path.dirname(os.path.dirname(__file__)), "bin/chai.exe")
SUITES_DIR = os.path.join(os.path.dirname(__file__), 'suites')
SUITE_DIRS = {'exercises': os.path.join(SUITES_DIR, 'exercises')}

def error(msg):
    print("[error]: " + msg)
    exit(1)

def run_test(test_name, test_dir):
    print(f"Running Test: `{test_name}`:")

    compile_proc = subprocess.run([EXEC_PATH, 'build', test_dir], stdout=subprocess.PIPE)
    if compile_proc.returncode != 0:
        print('Failed to compile test:')
        print(compile_proc.stdout.decode('UTF-8'))
        return

    exec_path = os.path.join(test_dir, 'out.exe')

    try:
        with open(os.path.join(test_dir, 'test.txt')) as f:
            case_n = 0
            passed = True
            p = None   

            def check_case():
                nonlocal passed
                nonlocal p

                if passed:
                    print(f"Test Case {case_n} Passed")
                else:
                    print(f"Test Case {case_n} Failed")
                    passed = True

                for line in p.stdout:
                    if line == "":
                        continue 

                    print("[test case {case_n} error]: Extra output" + line)

                p.kill()

            for line in f:
                if line.startswith('# case '):
                    if case_n != 0:            
                        check_case()

                    p = subprocess.Popen(exec_path, stdout=subprocess.PIPE, stdin=subprocess.PIPE) 
                    case_n = int(line[7:])
                elif case_n == 0:
                    error('Test missing case number')
                elif line.startswith('>>'):
                    p.stdin.write(line[3:] + '\n')
                elif line:
                    continue
                elif (pline := p.stdout.readline().decode('UTF-8')) != line:
                    print(f'[test case {case_n} error]: Expected {line} but got {pline}')
                    passed = False

            check_case()
    finally:
        if p:
            p.kill()

def run_test_dir(suite_name):
    suite_dir = SUITE_DIRS[suite_name]

    print(f"Running Test Suite: `{suite_name}`")
    print('-' * (22 + len(suite_name)))

    try:
        for dir in os.listdir(suite_dir):
            abs_dir_path = os.path.join(suite_dir, dir)

            if os.path.isdir(abs_dir_path):
                run_test(dir, abs_dir_path)
    except FileNotFoundError:
        error('Unable to open test case file')
    except TimeoutError:
        error('Executable timed out')
    except Exception as e:
        error(repr(e))

    print('\n')

if __name__ == '__main__':
    if len(sys.argv) != 2:
        error("expecting the name of a test suite to run")

    suite_name = sys.argv[1]
    
    if suite_name == 'all':
        for suite_name in SUITE_DIRS:
            run_test_dir(suite_name)
    elif suite_name in SUITE_DIRS:
        run_test_dir(suite_name)
    else:
        error(f"unknown suite name: {suite_name}")
