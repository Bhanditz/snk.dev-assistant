from typing import List, NamedTuple

import asyncio
import requests
from requests import Response
import re
from functools import total_ordering

from http_session_pool import RealWorldSessions
from proxy_provider import KDProxyProvider

RE_LINK = re.compile(r'\* \[.*?\]\(https://github\.com/(.*?)\) - .*')
KEY_STARS = 'watchers_count'
KEY_FORKS = 'forks_count'
KEY_ISSUES = 'open_issues'


@total_ordering
class RepoItem:
    text: str
    repository: str
    n_stars: int
    n_forks: int
    n_issues: int

    def __init__(self, text: str, repository: str, stats: dict) -> None:
        self.n_forks = stats.get(KEY_FORKS, 0)
        self.n_stars = stats.get(KEY_STARS, 0)
        self.n_issues = stats.get(KEY_ISSUES, 0)
        self.text = text
        self.repository = repository

    def __lt__(self, other):
        if isinstance(other, RepoItem):
            return self.n_stars < other.n_stars
        else:
            raise Exception('Invalid Parameter')

    def __repr__(self) -> str:
        return '|{}|{}|{}|{}|'.format(self.n_stars, self.n_forks, self.n_issues, self.text)


class ReadMeProcessor:
    def __init__(self, raw_url: str, loop) -> None:
        self.raw_url = raw_url
        self.loop = loop
        self.sessions = RealWorldSessions(num_sessions=100, proxy_provider=KDProxyProvider())

    def _get_readme(self) -> str:
        response: Response = requests.get(self.raw_url)
        try:
            return response.text
        finally:
            response.close()

    def _get_json(self, url: str) -> dict:
        with self.sessions.acquire() as s:
            response: Response = s.get(url)
            try:
                return response.json()  # type: dict
            finally:
                response.close()

    async def _get_github_repo(self, text: str, repo: str) -> RepoItem:
        stats: dict = await self.loop.run_in_executor(None, self._get_json,
                                                      f'https://api.github.com/repos/{repo}')
        return RepoItem(text, repo, stats)

    def _get_repo(self, line: str) -> str:
        matches = RE_LINK.findall(line)
        try:
            return matches[0]
        except IndexError:
            return ''

    async def _add_stats(self, readme: str) -> str:
        result = []
        lines = readme.split('\n')
        i = 0
        total = len(lines)
        buffer = []

        for line in lines:

            if line.startswith('##') and len(buffer):
                # new section
                results, _ = await asyncio.wait(buffer)
                result.append('|stars|forks|issues|description|')
                result.append('| --- | --- | --- | --- |')
                result.extend(sorted([el.result() for el in results], reverse=True))
                buffer.clear()

            else:
                repo: str = self._get_repo(line)
                if repo:
                    # get repository stats

                    buffer.append(self._get_github_repo(line, repo))
                else:
                    # not a repository
                    result.append(line)

            i += 1
            print(f'{i}/{total} complete')

        return "\n".join([str(e) for e in result])

    async def process(self) -> None:
        try:
            print('processing...')
            readme: str = self._get_readme()
            processed: str = await self._add_stats(readme)
            with open('README.md', 'w', encoding='utf-8') as f:
                f.write(f'this is a processed copy of {readme_url}\n\n')
                f.write(processed)

        finally:
            self.loop.stop()


if __name__ == '__main__':
    readme_url = 'https://raw.githubusercontent.com/avelino/awesome-go/master/README.md'
    loop = asyncio.get_event_loop()
    readme_processor = ReadMeProcessor(readme_url, loop)
    asyncio.ensure_future(readme_processor.process())
    loop.run_forever()
