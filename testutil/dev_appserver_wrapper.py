# -*- coding: utf-8 -*-

u"""Wraps dev_appserver and replaces options with environment variables.

* APPENGINE_DEV_APPSERVER_BASE
    * dev_appserver.py to launch. The one in paths are used if not specified.
* DEV_APPSERVER_API_PORT
    * API port to launch.
"""

import argparse
import os
import os.path
import sys


def findDevAppserver():
    devAppserver = os.environ.get('APPENGINE_DEV_APPSERVER_BASE')
    if devAppserver:
        return devAppserver

    for dir in os.getenv("PATH").split(os.pathsep):
        path = os.path.join(dir, 'dev_appserver.py')
        if os.path.exists(path):
            return path
    return None


def buildArguments():
    # Parse arguments to replace.
    parser = argparse.ArgumentParser()

    parser.add_argument(
        '--api_port',
        type=long,
        dest='api_port',
        default=0,
    )

    args, unknown = parser.parse_known_args()

    apiPort = os.environ.get('DEV_APPSERVER_API_PORT')
    if apiPort and apiPort.isdigit():
        args.api_port = long(apiPort)

    return [
        '--api_port={0}'.format(args.api_port),
    ] + unknown

if __name__ == '__main__':
    args = [
        sys.executable,
        findDevAppserver(),
    ] + buildArguments()

    os.execv(
        sys.executable,
        args,
    )
    sys.exit(1)
