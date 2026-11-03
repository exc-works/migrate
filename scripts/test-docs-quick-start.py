#!/usr/bin/env python3
"""Execute the website's actual SQLite snippets against a freshly built CLI."""

import json
import os
from pathlib import Path
import re
import sqlite3
import subprocess
import tempfile

repository = Path(__file__).resolve().parent.parent


def blocks(document, language):
    return re.findall(rf"```{language}\n(.*?)\n```", document, re.DOTALL)


with tempfile.TemporaryDirectory(prefix="migrate-docs-") as directory:
    workspace = Path(directory)
    binary = workspace / ("migrate.exe" if os.name == "nt" else "migrate")
    subprocess.run(["go", "build", "-o", str(binary), "./cmd/migrate"], cwd=repository, check=True)
    environment = dict(os.environ, PATH=str(workspace) + os.pathsep + os.environ.get("PATH", ""))
    for locale in ("", "zh"):
        document = (repository / "docs/site/pages" / locale / "quick-start.md").read_text()
        commands = blocks(document, "sh")
        assert len(commands) == 5, "Update this runner when the quick-start sequence changes"
        scenario = workspace / (locale or "en")
        scenario.mkdir()
        # The first block installs a released binary. Exercise the checked-out CLI instead.
        subprocess.run([str(binary), "--help"], check=True, capture_output=True)
        subprocess.run([str(binary), "version"], check=True, capture_output=True)
        subprocess.run(["bash", "-eu", "-c", commands[1]], cwd=scenario, env=environment, check=True)
        project = scenario / "migrate-demo"
        (project / "migration_config.json").write_text(blocks(document, "json")[0])
        subprocess.run(["bash", "-eu", "-c", commands[2]], cwd=project, env=environment, check=True)
        (project / "migrations/V202611030001__init_users.sql").write_text(blocks(document, "sql")[0])

        for command_block, expected in ((commands[3], "applied"), (commands[4], "pending")):
            subprocess.run(["bash", "-eu", "-c", command_block], cwd=project, env=environment, check=True)
            result = subprocess.run([str(binary), "status", "--output", "json"], cwd=project,
                                    check=True, capture_output=True, text=True)
            statuses = json.loads(result.stdout)
            assert len(statuses) == 1 and statuses[0]["status"] == expected, statuses
            with sqlite3.connect(project / "demo.sqlite") as database:
                table = database.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").fetchone()
                assert bool(table) == (expected == "applied")
        print(f"{locale or 'en'} quick start: applied and rolled back successfully")
