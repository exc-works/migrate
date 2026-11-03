#!/usr/bin/env python3
"""Check rendered Pages output, including local assets and heading fragments."""

from html.parser import HTMLParser
from pathlib import Path
import sys
from urllib.parse import unquote, urljoin, urlsplit


class Page(HTMLParser):
    def __init__(self, html):
        super().__init__(convert_charrefs=True)
        self.links = []
        self.ids = set()
        self.has_heading = False
        self.feed(html)

    def handle_starttag(self, tag, attributes):
        attributes = dict(attributes)
        if "id" in attributes:
            self.ids.add(attributes["id"])
        self.has_heading |= tag == "h1"
        for attribute in ("href", "src"):
            if attributes.get(attribute):
                self.links.append(attributes[attribute])


def check_site(output):
    site_url = "https://exc-works.github.io/migrate/"
    site = urlsplit(site_url)
    output = output.resolve()
    pages = {path: Page(path.read_text()) for path in output.rglob("*.html")}
    if not pages:
        raise SystemExit("No HTML pages found; build the documentation first")
    errors = []
    for path, page in pages.items():
        relative_path = path.relative_to(output).as_posix()
        if relative_path != "404.html" and not page.has_heading:
            errors.append(f"{relative_path}: no rendered h1; page rendering may have failed")
        for link in page.links:
            target = urlsplit(urljoin(urljoin(site_url, relative_path), link))
            if target.netloc != site.netloc:
                continue
            if not target.path.startswith(site.path):
                errors.append(f"{relative_path}: outside Pages base: {link}")
                continue
            relative_target = unquote(target.path[len(site.path):])
            candidates = (output / relative_target, output / (relative_target + ".html"),
                          output / relative_target / "index.html")
            destination = next((candidate.resolve() for candidate in candidates
                                if candidate.resolve().is_relative_to(output) and candidate.is_file()), None)
            if destination is None:
                errors.append(f"{relative_path}: missing file: {link}")
            elif target.fragment and destination in pages:
                if unquote(target.fragment) not in pages[destination].ids:
                    errors.append(f"{relative_path}: missing fragment: {link}")
    if errors:
        raise SystemExit("\n".join(errors))
    print(f"Checked {len(pages)} rendered pages, local links, assets and heading fragments")


if __name__ == "__main__":
    default_output = Path(__file__).resolve().parent.parent / "docs/site/.vitepress/dist"
    check_site(Path(sys.argv[1]) if len(sys.argv) > 1 else default_output)
