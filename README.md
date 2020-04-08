# Beast Changelog Action

> This changelog action is a beast :)

Maybe not entirely true. This action aims to do a few things:

* Give credit where credit is due by linking to pull request author profiles
* Differentiate between direct commits and pull requests
* Ignore merge commits
* Allow for template-based customization of outputs (for example if you target AsciiDoc rather than Markdown)
* Group (and optionally exclude) commits from the changelog based on patterns
* Don't assume tag formats or anything else, just operate linearly _from_ one hash and _to_ another

_Beast Changelog Action_ is built on top of [jimschubert/changelog](https://github.com/jimschubert/changelog), which could be wrapped similarly to this action if you need to target another CI tool.

---

The changelog format following this repository's [full example](.github/workflows/example.yml) will resemble:

```markdown
## v1

### Features

* [426dbf4915](https://github.com/jimschubert/beast-changelog-action/commit/426dbf49151c269024a7c3ca4af83a49909a5b7b) Update action.yml to use image on dockerhub ([jimschubert](https://github.com/jimschubert))

### Other

* [0fc811905d](https://github.com/jimschubert/beast-changelog-action/commit/0fc811905d02aa316cf0ad8226bb8fc57bc181fb) doc: Create README for action ([jimschubert](https://github.com/jimschubert))

<em>For more details, see <a href="https://github.com/jimschubert/beast-changelog-action/compare/v0.1...v1">v0.1..v1</a></em>
```

---

## Inputs

### `GITHUB_TOKEN`

**Required** Default: `${{github.token}}`

GitHub token used to access the repository defined in the GITHUB_REPOSITORY input.

It is recommended to [create a new personal access token](https://github.com/settings/tokens/new) with the least permissions (e.g. public_repo).
Using a service account for the GitHub Token is also highly recommended.

[Learn more about using secrets](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/creating-and-using-encrypted-secrets)

### `GITHUB_REPOSITORY`

**Required** Default: `${{github.repository}}`

The target github repo in the format owner/repo

### `CONFIG_LOCATION`

**Required** Default: `.github/changelog.json`

The file location to the changelog configuration.

See [jimschubert/changelog](https://github.com/jimschubert/changelog) for schema and further details.

### `OUTPUT`

**Required** Default: `.github/CHANGELOG.md`

The output file where the changelog will be written.

The file is created and appended, but _not_ committed back to the repository.

It is recommended to add a post-processing step in your workflow to prepend to an existing changelog.

### `FROM`

**Required**

The beginning tag from which to generate the changelog.

This can be queried on an unshallow-ed repository with:

```
git describe --tags --abbrev=0 --match 'v*' HEAD~
```

See also [jimschubert/query-tag-action](https://github.com/jimschubert/query-tag-action).

### `TO`

**Required**

The ending tag until which the changelog should be generated.

This can be queried on an unshallow-ed repository with:

```
git describe --tags --abbrev=0 --match 'v*' HEAD
```

See also [jimschubert/query-tag-action](https://github.com/jimschubert/query-tag-action).

## Outputs

The action itself does not output any arguments. Your settings _may_ result in an artifact carried between steps.

## Usage

### Define `.github/changelog.json`

This action supports JSON and YAML 1.1 configuration files, but defaults to JSON.

The schema is as follows:

```json5
{
  // "commits" or "prs", defaults to commits. "prs" will soon allow for resolving labels 
  // from pull requests
  "resolve": "commits",

  // "asc" or "desc", determines the order of commits in the output
  "sort": "asc",
  
  // GitHub user or org name
  "owner": "jimschubert",  
   
  // Repository name
  "repo": "changelog",

  // Enterprise GitHub base url
  "enterprise": "https://ghe.example.com",

  // Path to custom template following Go Text template syntax
  "template": "/path/to/your/template",

  // Group commits by headings based on patterns supporting Perl syntax regex or plain text
  "groupings": [
    { "name":  "Contributions", "patterns":  [ "(?i)\\bfeat\\b" ] }
  ],

  // Exclude commits based on this set of patterns or texts
  // (useful for common maintenance commit messages)
  "exclude": [
    "^(?i)release\\s+\\d+\\.\\d+\\.\\d+",
    "^(?i)minor fix\\b",
    "^(?i)wip\\b"
  ],
   
  // Prefers local commits over API. Requires executing from within a Git repository.
  "local": false,
 
  // Processes UP TO this many commits before processing exclusion/inclusion rules. Defaults to size returned from GitHub API.
  "max_commits": 250
}
```

Regex patterns for groupings and exclusions must be escaped according to the [ECMA-404 JSON Data Interchange Syntax](https://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf) specification.

An example YAML file might look like:

```yaml
%YAML 1.1
---
resolve: commits
owner: jimschubert
repo: ossify
groupings:
  - name: feature
    patterns: ['^a','\bb$']
  - name: bug
    patterns: ['cba','\b\[f\]\b']
exclude:
  - wip
  - help wanted
enterprise: 'https://ghe.example.com'
template: /path/to/template
sort: asc
max_commits: 199
local: true
``` 

### Create a workflow

A [full example](.github/workflows/example.yml) is available in this repository.

At a minimum, your workflow _must_:

* Checkout the git repository (e.g. via `actions/checkout@v2`)
* Unshallow the git repository
* Query an earlier tag
* Query the later tag
* Include the Beast Changelog Action

#### Unshallow

GitHub Actions create a [shallow](https://www.git-scm.com/docs/shallow) clone by default. This reduces traffic and handles most build-related use cases.

We need to "unshallow" the repository in order to access tag information. *If you forget to unshallow, the changelog step will fail.*

To unshallow, simply run `git fetch --prune --unshallow`

#### Query the **FROM** Tag

It doesn't really matter how you query the starting tag. This value could be hard-coded and updated manually, pulled via GitHub's Releases API, or queried via rules as an individual step.

The example in this repository uses one of my other actions, [jimschubert/query-tag-action@v1](https://github.com/jimschubert/query-tag-action), to accept some tag rules and output the found tag:

```yaml
- name: Find Last Tag
  id: last
  uses: jimschubert/query-tag-action@v1
  with:
    include: 'v*'
    exclude: '*-rc*'
    commit-ish: 'HEAD~'
    skip-unshallow: 'true'
```

This action will output the found tag (or fail if no tag is found). It can later be referenced according to the `id` as `${{steps.last.outputs.tag}}`.

The above query-tag-action example is simply a wrapper around the command line:

```bash
git describe --tags --abbrev=0 --match 'v*' --exclude '*-rc*' HEAD~
```

Note that the above query action doesn't have any way to know if the last tag was a released tag or not.

If you want to generate a changelog against only those tags which have been released via GitHub Releases, you could use `curl` and `jq` to pull from GitHub Releases and take the last version. As an example:

```bash
function latest.tag {
  local uri="https://api.github.com/repos/${1}/releases"
  local ver=$(curl -s ${uri} | jq -r 'first(.[]|select(.prerelease==false)).tag_name')
  echo $ver
}

echo "$(latest.tag ${GITHUB_REPOSITORY})"
```

The `GITHUB_REPOSITORY` variable is available to all GitHub Actions. See GitHub Action documentation on [using environment variables](https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables) for more details.

#### Query the **TO** Tag

The latest tag can also be queried using [jimschubert/query-tag-action@v1](https://github.com/jimschubert/query-tag-action):

```yaml
- name: Find Current Tag
id: current
uses: jimschubert/query-tag-action@v1
with:
  include: 'v*'
  exclude: '*-rc*'
  commit-ish: '@'
  skip-unshallow: 'true'
```

This action will output the found tag (or fail if no tag is found). It can later be referenced according to the `id` as `${{steps.current.outputs.tag}}`.

If your workflow runs on the [release event](https://help.github.com/en/actions/reference/events-that-trigger-workflows#release-event-release), you may also manipulate the `GITHUB_REF` variable:

```bash
latest=${GITHUB_REF#refs/tags/}
```

Note that the above will always take the tag applied for the current release build and doesn't really allow for filtering as with the above action. If you apply multiple tags to the same commit (e.g. `v1.0.0-rc`, `v1.0.0-prerelease`), the workflow run will result in the tag which triggered that run.

#### Include the Beast Changelog Action 

The action may be defined with full options (FROM and TO options assume query-tag-action examples from above):

```yaml
- name: Create Changelog
  id: changelog
  uses: jimschubert/beast-changelog-action@v1
  with:
    GITHUB_TOKEN: ${{github.token}}
    GITHUB_REPOSITORY: ${{github.repository}}
    CONFIG_LOCATION: .github/changelog.json
    FROM: ${{steps.last.outputs.tag}}
    TO: ${{steps.current.outputs.tag}}
    OUTPUT: .github/CHANGELOG.md
```

The `GITHUB_TOKEN`, `GITHUB_REPOSITORY`, and `CONFIG_LOCATION` inputs are optional. When should you change these?

* You should consider creating a secret using a service account whenever using a third-party action, as the `github.token`/`GITHUB_TOKEN` provided by default to actions has full account access. This action requires only `public_repo` access.
* You really only need to modify `GITHUB_REPOSITORY` if you're orchestrating changelogs from a repository which is different from your target repository.
* You should update `CONFIG_LOCATION` if you store your changelog configuration somewhere other than the default (for instance, if you decide to use YAML).

The `FROM` and `TO` options are required and have no defaults.

The `OUTPUT` option may be omitted, but `.github/CHANGELOG.md` may not be a desirable location. This file is created and appended to.

*This action does not support prepending to an existing file.* Please add another step to prepend `.github/CHANGELOG.md` to your target file.

#### Full Workflow Example

The following example workflow could be used to copy/paste changelogs into a Release created in GitHub:

```yaml
name: Dump Changelog
on: [release]

jobs:

  changelogger:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Unshallow
        run: git fetch --prune --unshallow
      - name: Find Last Tag
        id: last
        uses: jimschubert/query-tag-action@v1
        with:
          include: 'v*'
          exclude: '*-rc*'
          commit-ish: 'HEAD~'
          skip-unshallow: 'true'
      - name: Find Current Tag
        id: current
        uses: jimschubert/query-tag-action@v1
        with:
          include: 'v*'
          exclude: '*-rc*'
          commit-ish: '@'
          skip-unshallow: 'true'
      - name: Create Changelog
        id: changelog
        uses: jimschubert/beast-changelog-action@v1
        with:
          FROM: ${{steps.last.outputs.tag}}
          TO: ${{steps.current.outputs.tag}}
      - name: View Changelog
        run: cat .github/CHANGELOG.md
```

Inspect the output of the `View Changelog` step.

## License

This project is [licensed](./LICENSE) under Apache 2.0.
