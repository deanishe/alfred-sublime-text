#!/usr/bin/env python
# encoding: utf-8
#
# Copyright © 2014 deanishe@deanishe.net
#
# MIT Licence. See http://opensource.org/licenses/MIT
#
# Created on 2014-06-12
#

"""
"""

from __future__ import print_function, unicode_literals

import sys
import os
import subprocess

from workflow import Workflow, ICON_WARNING

VERSION = '1.0'

# Log filter scores (and rules)
DEBUG = True
# How long to cache the list of projects for (in seconds)
CACHE_DURATION = 20

# Will be set when the Workflow object is instantiated
log = None
decode = None


def find_projects():
    """Return a list of Sublime Text project files"""
    cmd = ['mdfind', '-name', '.sublime-project']
    output = subprocess.check_output(cmd)
    lines = [s.strip() for s in decode(output).split('\n') if s.strip()]
    log.debug('{} projects found'.format(len(lines)))
    return lines


def main(wf):
    query = None
    args = wf.args
    if len(args):
        query = args[0]

    projects = wf.cached_data('projects', find_projects, max_age=15)

    if query:
        projects = wf.filter(query, projects,
                             key=lambda p: os.path.basename(p),
                             include_score=DEBUG,
                             min_score=20)

        if DEBUG:  # Show scores, rules
            for (path, score, rule) in projects:
                log.debug('{:0.2f} [{}] {}'.format(score, rule, path))

            projects = [t[0] for t in projects]

    if not projects:
        wf.add_item('No matches found',
                    'Try a different query',
                    icon=ICON_WARNING)

    home = os.getenv('HOME')
    for path in projects:
        wf.add_item(os.path.basename(path).replace('.sublime-project', ''),
                    path.replace(home, '~'),
                    arg=path,
                    uid=path,
                    valid=True,
                    icon='document_icon.png')

    wf.send_feedback()


if __name__ == '__main__':
    wf = Workflow()
    log = wf.logger
    decode = wf.decode
    sys.exit(wf.run(main))
