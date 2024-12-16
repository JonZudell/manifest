# Customs

Customs is a Go application that is designed to lint pull requests and diffs
using configurable rules. It is language agnostic, passing the relevant pull
request and diff information to scripts via JSON while using the resulting stdout JSON to comment on the PR/diffs, fail the build, etc.

## Installing Customs

TODO

## Usage

### Getting import JSON to test scripts

Since customs work primarily through piping stdin and stdout, you'll need to generate the relevant JSON to pass to scripts utilizing `customs`. To get JSON usable for testing or running customs inspectors, you can pass `--only-import-json` to bypass running the configured scripts and return only the import JSON that would be passed to the inspectors.

```sh
$ cat my.diff | go run cmd/customs/main.go inspect --only-import-json
```

Which should result in output like:

```json
{
  "pullProvided": false,
  "diff": {
    "changed": [
      "app/jobs/greeter_job.rb"
    ],
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

## Writing a script

Customs scripts are extremely simple. They are passed JSON as stdin and are expected to output JSON to stdout that looks like the following:

Stdin:

```json
{
  "pullTitle": "Update job",
  "pullDescription": "Update the greeter to accept a name",
  "pullProvided": true,
  "diff": {
    "changed": [
      "app/jobs/greeter_job.rb"
    ],
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

