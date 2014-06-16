#!/usr/bin/python
# encoding: utf-8
#
# Copyright © 2014 deanishe@deanishe.net
#
# MIT Licence. See http://opensource.org/licenses/MIT
#
# Created on 2014-06-16
#

"""
Find .sublime-project file on system via a variety of methods.

The primary method is `mdfind`, but this won't return files in hidden
directories.

Use `find` on user's custom search directories, which is very slow.

Also ask `locate`, but it's database may be up to a week old, which
isn't much help (it's also slow). Still, it presents a decent fallback
if the directories in question can't be searched using `find` due to size.

"""

from __future__ import print_function, unicode_literals

import sys
import os
import subprocess
from time import time
from fnmatch import fnmatch

from workflow import Workflow


LOCATE_DB = '/var/db/locate.database'


log = None
decode = None


def run_command(cmd):
    """Run the command and return ``Popen`` object"""
    return subprocess.Popen(cmd, stdout=subprocess.PIPE)


def main(wf):
    start = time()
    procs = []

    # Get cached paths
    paths = wf.cached_data('projects', max_age=0)

    if not paths:
        paths = set()
    else:
        paths = set(paths)

    # Start `locate` first, as it takes a long time. Only grab results
    # from `locate` if its database has been updated since it was last
    # read
    if os.path.exists(LOCATE_DB):
        locatedb_cached = wf.settings.get('locatedb_cached', 0)
        locatedb_updated = os.stat(LOCATE_DB).st_mtime
        if locatedb_updated > locatedb_cached:
            log.debug('Searching `located` database ...')
            procs.append(run_command(['locate', '*.sublime-project']))
            wf.settings['locatedb_cached'] = time()

    # Search user-defined directories with `find`
    directories = wf.settings.get('search_directories', [])
    for path in directories:
        path = os.path.expanduser(path)
        if not os.path.exists(path) or not os.path.isdir(path):
            continue
        log.debug('Searching {} with `find` ...'.format(path))
        procs.append(run_command(['find', path, '-type', 'f',
                                  '-name', '*.sublime-project']))

    # mdfind
    log.debug('Searching with `mdfind` ...')
    procs.append(run_command(['mdfind', '-name', '.sublime-project']))

    while len(procs):
        proc = procs.pop(0)
        if proc.poll() is None:
            procs.append(proc)
            continue
        output = proc.communicate()[0]
        paths.update([s.strip() for s in
                      decode(output).split('\n') if s.strip()])

    projects = []
    exclude_patterns = wf.settings.get('excludes', [])
    for path in paths:
        # Exclude paths that don't exist. This is important, as `locate`
        # returns paths that may have long since been deleted or are on
        # drives that are currently not connected
        if not os.path.exists(path):
            continue
        valid = True
        # Exclude results based on globbing patterns
        for pat in exclude_patterns:
            if fnmatch(path, pat):
                log.debug('Excluded [{}] {}'.format(pat, path))
                valid = False
                break
        if valid:
            projects.append(path)

    # Save data to cache
    wf.cache_data('projects', sorted(projects))

    log.debug('{} projects found in {:0.2f} seconds'.format(
              len(paths), time() - start))

if __name__ == '__main__':
    wf = Workflow()
    log = wf.logger
    decode = wf.decode
    sys.exit(wf.run(main))
