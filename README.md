# Customs

Customs is a Go application that is designed to lint pull requests and diffs
using configurable rules. It is language agnostic, passing the relevant pull
request and diff information to scripts via JSON while using the resulting
stdout JSON to comment on the PR/diffs, fail the build, etc.

## Installing Customs

Run `go install github.com/blakewilliams/customs/cmd/customs` or clone+build from source.

## Usage

The primary usage of `customs` is via `customs inspect`, which can be configured directly in the CLI (see `customs inspect help`) or a configuration file in your root directory called `customs.config.yaml`:

```yaml
# Sample YAML config
customs:
  concurrency: 2 # How many inspectors to run at once
  formatter: pretty # The formatter to use
  inspectors: # The inspector scripts to run and report on
    feature_flags:
      command: "script/feature-flag-inspector"
    rails_job_perform:
      command: "script/job-perform-inspector"
```

Then you can run `git diff main | customs inspect` which will run each of the provided
inspectors in the provided config. Arguments provided in the config can be
overridden using the CLI flags ( see `customs inspect help`).

## Writing a custom inspector

Customs inspectors can be written in any language since they effectively accept
JSON as stdin, and output JSON in stdout so `customs` can output it
appropriately. The following JSON is the expected format:

Stdin:

```json
{
  "pullTitle": "Update job",
  "pullDescription": "Update the greeter to accept a name",
  "pullProvided": true,
  "diff": {
    "changed": ["app/jobs/greeter_job.rb"],
    "deleted": [],
    "renamed": [],
    "new": [],
    "copied": [],
    "files": {
      "app/jobs/greeter_job.rb": {
        "operation": "change",
        "new_name": "app/jobs/greeter_job.rb",
        "old_name": "app/jobs/greeter_job.rb",
        "left": [
          {
            "lineno": 4,
            "content": "  def perform\n"
          }
        ],
        "right": [
          {
            "lineno": 4,
            "content": "  def perform(name)\n"
          }
        ]
      }
    }
  }
}
```

Stdout:

```json
{
  "error": "",
  "comments": [
    {
      "file": "app/jobs/greeter_job.rb",
      "line": 4,
      "text": "You have modified an ActiveRecord job's arguments. In order to avoid job failures please read and follow X documentation.",
      "severity": "Warn"
    }
  ]
}
```

Comments are then output to stdout or posted to Pull Requests. The format of comments should be:

```json
{
  "file": "app/jobs/greeter_job.rb",         // optional file, missing file+line comments top-level
  "line": 4,                                 // optional line number
  "text": "don't do that because...!",       // The text to output
  "severity": "Warn"                         // The severity of the violation. Can be one of Info, Warn, or Error.
}
```


See also the `Result` struct in `result.go` for more details on the expected output format and the `Import` struct in `customs.go` for the expected inputs.

### Getting import JSON to test scripts

Since customs work primarily through piping stdin and stdout, you'll need to generate the relevant JSON to pass to scripts utilizing `customs`. To get JSON usable for testing or running customs inspectors, you can pass `--only-import-json` to bypass running the configured scripts and return only the import JSON that would be passed to the inspectors.

```sh
$ cat my.diff | customs inspect --only-import-json
```

Which should result in output like:

```json
{
  "pullProvided": false,
  "diff": {
    "changed": ["app/jobs/greeter_job.rb"],
    "deleted": [],
    "renamed": [],
    "new": [],
    "copied": [],
    "files": {
      "app/jobs/greeter_job.rb": {
        "operation": "change",
        "new_name": "app/jobs/greeter_job.rb",
        "old_name": "app/jobs/greeter_job.rb",
        "left": [
          {
            "lineno": 4,
            "content": "  def perform\n"
          }
        ],
        "right": [
          {
            "lineno": 4,
            "content": "  def perform(name)\n"
          }
        ]
      }
    }
  }
}
```

Using this output, you can pipe the JSON directly into your inspector script:

```sh
$ cat my.diff | customs inspect --only-import-json | my-inspector
```
