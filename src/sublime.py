#!/usr/bin/python
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
import argparse
import subprocess

from workflow import (Workflow,
                      ICON_WARNING, ICON_INFO, ICON_SETTINGS, ICON_SYNC)
from workflow.background import run_in_background, is_running

VERSION = '2.0'

# Location of `locate` database
LOCATE_DB = '/var/db/locate.database'

# Log filter scores (and rules)
DEBUG = False
# How long to cache the list of projects for (in seconds)
MDFIND_INTERVAL = 15

DEFAULT_SETTINGS = {
    'locatedb_cached': 0,
    'excludes': ['/Applications/*.app/*'],
    'search_directories': ['~/.config', '~/.dotfiles'],
}

# Will be set when the Workflow object is instantiated
log = None


def do_edit_config(wf):
    """Open `settings.json` in default editor"""
    subprocess.call(['open', wf.settings_path])


def do_update(wf):
    """Delete cached results to force update on next call"""
    # Force reloading of results from `locate`
    wf.settings['locatedb_cached'] = 0
    # Delete cached projects list
    wf.cache_data('projects', None)
    print('Forced project list update')


def do_help(wf):
    """Open help file in browser"""
    subprocess.call(['open', wf.workflowfile('Help.html')])


def do_config(wf):
    """Show configuration"""
    wf.add_item('Edit Configuration',
                'Open `settings.json` in your default editor.',
                valid=True,
                arg='edit',
                icon=ICON_SETTINGS)

    if not os.path.exists(LOCATE_DB):
        wf.add_item('`locate` is not active on your system',
                    'Action this item to learn about `locate`',
                    valid=True,
                    arg='help',
                    icon=ICON_WARNING)
    else:
        wf.add_item('Force update',
                    "Use this if you're not seeing results you expect.",
                    valid=True,
                    arg='update',
                    icon=ICON_SYNC)

    # Custom search directories
    dirs = wf.settings.get('search_directories', [])
    if not dirs:
        wf.add_item('No custom search directories set',
                    'Choose "Edit Configuration" to add some.',
                    icon=ICON_INFO)

    home = os.getenv('HOME')
    for path in dirs:
        path = os.path.abspath(os.path.expanduser(path))
        short_path = path.replace(home, '~')
        if os.path.exists(path):
            msg = 'Also search in this directory using `find`.'
            icon = path
            icontype = 'fileicon'
        else:
            msg = 'This directory does not exist. Consider removing it.'
            icon = ICON_WARNING
            icontype = None
        wf.add_item(short_path,
                    msg,
                    icon=icon,
                    icontype=icontype)

    # Exclude patterns
    excludes = wf.settings.get('excludes', [])
    if not excludes:
        wf.add_item('No exclude patterns set',
                    'Choose "Edit Configuration" to add some.',
                    icon=ICON_INFO)

    for pat in excludes:
        wf.add_item(pat,
                    'Exclude files matching this globbing pattern.',
                    icon='Exclude.png')

    wf.send_feedback()


def main(wf):
    parser = argparse.ArgumentParser()
    parser.add_argument('-a', '--action',
                        dest='action',
                        help='Perform action')
    parser.add_argument('query', help='What to filter projects by',
                        nargs='?', default=None)
    args = parser.parse_args(wf.args)
    query = args.query

    if args.action:
        if args.action == 'config':
            return do_config(wf)
        elif args.action == 'help':
            return do_help(wf)
        elif args.action == 'update':
            return do_update(wf)
        elif args.action == 'edit':
            return do_edit_config(wf)
        else:
            raise ValueError('Unknown action : {}'.format(args.action))

    # Create default settings
    if not os.path.exists(wf.settings_path):
        for key in DEFAULT_SETTINGS:
            wf.settings[key] = DEFAULT_SETTINGS[key]

    # Load cached data if it exists. If it's out-of-date, we'll take
    # care of that directly
    projects = wf.cached_data('projects', None, max_age=0)

    # Update cache if it's out-of-date
    if not wf.cached_data_fresh('projects', max_age=MDFIND_INTERVAL):
        cmd = ['/usr/bin/python', wf.workflowfile('update_cache.py')]
        run_in_background('update', cmd)

    # Show update message if cache is empty
    if is_running('update') and not projects:
        wf.add_item('Generating list of Sublime Text projects',
                    valid=False,
                    icon=ICON_INFO)

    if query and projects:
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
    else:
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
    sys.exit(wf.run(main))
