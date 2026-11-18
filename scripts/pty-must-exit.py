#!/usr/bin/env python3
"""Run a command under a pty. Exit 3 if it is still running after the deadline."""

import os
import pty
import select
import sys
import time


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: pty-must-exit.py <binary> [args...]", file=sys.stderr)
        return 2

    pid, fd = pty.fork()
    if pid == 0:
        try:
            os.execv(sys.argv[1], sys.argv[1:])
        except OSError:
            os._exit(127)

    deadline = time.time() + 8.0
    status = None
    while time.time() < deadline:
        try:
            waited, got = os.waitpid(pid, os.WNOHANG)
        except ChildProcessError:
            return 3
        if waited == pid:
            status = got
            break
        ready, _, _ = select.select([fd], [], [], 0.2)
        if not ready:
            continue
        try:
            os.read(fd, 4096)
        except OSError:
            pass

    if status is None:
        try:
            os.kill(pid, 9)
        except OSError:
            pass
        os.waitpid(pid, 0)
        print("pty-must-exit: timed out", file=sys.stderr)
        return 3

    if os.WIFEXITED(status):
        return os.WEXITSTATUS(status)
    return 3


if __name__ == "__main__":
    sys.exit(main())
