# Skills used by Edda's game master

These folders are the game-mastering and fiction skills Edda's GM can load
during a turn through the `load_skill` and `read_skill_file` tools
(`internal/llm/skills.go`). Only their Markdown is copied; the scripts and data
files of the originals are not used.

## Origin and license

The skills were written by jwynia and published in
[jwynia/agent-skills](https://github.com/jwynia/agent-skills), where each
skill's front matter declares `license: MIT` and `author: jwynia`. They were
renamed and restructured in patstack, the owner's skill library; these copies
come from patstack revision `be697cd` on 2026-11-14. Changes made in patstack and
in this repository are under Edda's license (GPL-3.0-or-later). The original
material remains under the MIT license below.

```text
MIT License

Copyright (c) jwynia (https://github.com/jwynia/agent-skills)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

The upstream repository has no repository-wide license file or copyright line;
the notice above names the author as the skills' front matter does.

## Updating

Copy the Markdown files again from patstack's `library/skills/<name>` and
update the revision above.
