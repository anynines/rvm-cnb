# RVM Cloud Native Buildpack in Go

The RVM Cloud Native Buildpack in Go installs RVM and a Ruby version in OCI image. It has been created for usage in conjunction with the [Rails builder](https://github.com/avarteqgmbh/rails-builder-cnb).

## Functionality

1. The RVM CNB installs RVM into its own layer. The version of RVM to be installed can be configured in [buildpack.toml](buildpack.toml).
1. It also install a version of Ruby using RVM. The version to be installed is selected as follows (in order of precedence):
    1. If there is a `.ruby-version` file, its contents are used to select the Ruby version.
    1. If there is a file called `Gemfile`, then the string "ruby \<version string\>" is searched for in this file and if it exists, the given Ruby version is selected.
    1. If there is a file called `Gemfile.lock`, then the string "RUBY VERSION" is searched for in this file and if it exists, the contents of the next line is used to select the Ruby version.
    1. If none of the files specified above exists, then the Ruby version specified in [buildpack.toml](buildpack.toml) will be selected. The variable that specifies the default Ruby version is called `default_ruby_version`.

## Dependencies

This CNB installs the [Node CNB](https://github.com/paketo-buildpacks/node-engine) as a dependency in the build and launch layers. Currently, the default version of Node installed is the latest `12.*` version.

## TODO

1. Tests have to be written.
