#!/usr/bin/env python3
"""Run a command under a pty and answer one confirmation line. Test helper only."""
import os
import pty
import select
import sys
import time


def main() -> int:
    if len(sys.argv) < 3:
        print("usage: install-pty-test.py <binary> <phrase> [args...]", file=sys.stderr)
        return 3

    binary = sys.argv[1]
    phrase = sys.argv[2]
    args = sys.argv[3:]

    pid, fd = pty.fork()
    if pid == 0:
        try:
            os.execv(binary, [binary, *args])
        except OSError:
            sys.exit(127)

    buf = b""
    sent = False
    end = time.time() + 15.0

    while time.time() < end:
        rlist, _, _ = select.select([fd], [], [], 0.2)
        if not rlist:
            continue
        chunk = os.read(fd, 4096)
        if not chunk:
            break
        buf += chunk
        want = ("type " + (args[-1] if args else "")).encode()
        if not sent and want in buf:
            os.write(fd, (phrase + "\n").encode())
            sent = True

    while True:
        rlist, _, _ = select.select([fd], [], [], 0.2)
        if not rlist:
            break
        chunk = os.read(fd, 4096)
        if not chunk:
            break
        buf += chunk

    _, status = os.waitpid(pid, 0)
    sys.stdout.buffer.write(buf)
    sys.stdout.buffer.flush()
    if os.WIFEXITED(status):
        return os.WEXITSTATUS(status)
    return 3


if __name__ == "__main__":
    sys.exit(main())
