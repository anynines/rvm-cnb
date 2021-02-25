# RVM Cloud Native Buildpack

The RVM Cloud Native Buildpack installs RVM and a Ruby version in an OCI image. It has been created for usage in conjunction with the [Rails builder](https://github.com/avarteqgmbh/rails-builder-cnb).

## Functionality

1. The RVM CNB installs RVM into its own layer. The version of RVM to be installed can be configured in [buildpack.toml](buildpack.toml).
1. `Gemfile` and `Gemfile.lock` files must exist in the application directory in order to RVM CNB
can be able to start DETECTION phase. 
1. It also installs a version of Ruby using RVM. The version to be installed is selected as follows (in order of precedence, the method listed highest wins):
    1. If there is a called `buildpack.yml` in the application directory, it may specify a ruby version. See below to learn possible keys in the buildpack.yml file.
    1. If there is a file called `Gemfile.lock`, then the string "RUBY VERSION" is searched within this file and if it exists, the contents of the next line is used to select the Ruby version.
    1. If there is a file called `Gemfile`, then the string "ruby \<version string\>" is searched within this file and if it exists, the given Ruby version is selected.
    1. If there is a `.ruby-version` file, its contents are used to select the Ruby version.
    1. If none of the files specified above exists, then the Ruby version specified in [buildpack.toml](buildpack.toml) will be selected. The variable that specifies the default Ruby version is called `default_ruby_version`.

### buildpack.yml

A buildpack.yml may specify the following keys. If buildpack.yml specifies a Ruby version, it will have priority over all other Ruby version sources.

```yaml
rvm:
  rvm_version: 1.29.10
  ruby_version: 2.6.1
  node_version: 10.*
  require_node: true
```

## Dependencies

This CNB installs the [Node CNB](https://github.com/paketo-buildpacks/node-engine) as a dependency in the build and launch layers. Currently, the default version of Node installed is the latest `12.*` version.

## TODO

1. Refactor this CNB and remove Bundler specific functionality
1. Only leave functionality directly related to the installation or configuration of RVM and a particular Ruby version
