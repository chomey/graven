# graven

Graven is a build management tool for Go projects. It takes light
cues from projects like Maven and Leiningen, but given Go's much
simpler environment and far different take on dependency management, 
little is shared beyond the goals.

These shared goals are as follows.

## Providing an easy, uniform build system

Go projects are often built with makefiles or bash scripts. As much 
as I like Make, given Go's cross platform capabilities, Make is 
a poor experience for Windows developers. In addition, Make always
suffers from a lack of consistency, and also carries a bit of 
legacy baggage that is overhead that is simply not necessary 
anymore. 

Graven offers a single, consistent artifact: `project.yaml` that
can be used to build a project on Mac, Linux or Windows in a 
consistent and easy way. It also supports creating consistent
archives of deployable executables and any required resources.

The Graven command line interface offers a simple lifecycle 
similar to that which is often implemented in makefiles or
bash scripts: `clean`, `build`, `test`, `package`, `release`.

## Providing project information and guidelines for best practices

For all the greatness that is Go, there is a major gap in practices
surrounding versioning and dependency management. Vendoring has slightly
improved the dependency management, but lacks a single consistent tool.
Furthermore, builds aren't easily repeatable and versions are usually
based on commit hashcodes, rather than intelligently selected semantic
versions that describe capabilities, compatibility and bug fixes. 

Graven supports, automate and encourages proper semantic versioning and 
can freeze vendor dependencies to ensure repeatable builds are possible. 
Graven is opinionated about vendoring tools, and has chosen Govendor as 
its standard. However, it may support other vendoring tools in the future, 
and will embrace any standard tools that eventually come from the Go
project.


## Where things differ

While Graven takes queues from Maven and Leiningen, it also casts out 
the annoying, verbose and repetitive aspects that most developers
agree weigh Maven down. 

So Graven embraces:

* A much simpler build artifact based on light YAML
* Batteries included, no plugins - none are even supported yet, 
which is considered a good thing for now

# Dependencies

Graven currently requires the following tools to be on your path:

```
go - the Go build tool, used to compile and test your application.
git - used during the release process to validate the state of your repo, and tag your repo.
govendor - used during the freeze/unfreeze process to lock in your dependencies.
```

Of course if you don't plan to use the `release` command or the `freeze` and `unfreeze` commands, you
can still use `graven` just for building, testing and packaging, and thus would only require the
`go` tool. 

# Usage

```
$ ./graven
NAME:
   graven - A build automation tool for Go.

USAGE:
   graven [global options] command [command options] [arguments...]

VERSION:
   0.3.0

COMMANDS:
     build     Builds the current project
     info      Prints the known information about a project
     clean     Cleans the target directory and its contents
     package   Produces packaged archive for deployment
     bump      Manage the version (major, minor, patch) and clear or set qualifier (e.g. DEV)
     test      Finds and runs tests in this project
     freeze    Freezes vendor dependencies to avoid having to check in source
     unfreeze  Unfreezes vendor dependencies
     init      Initializes a project directory
     release   Releases artifacts to repositories
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```
# Workflow Example

Whether your starting an entirely new project, or working with existing source, the 
workflow should be the same. 

Not all of these steps are always necessary. See below for a description of implied 
workflow dependencies.

```

# Once per project, run the init command and modify
# default project.yaml with relevant names and repos etc.
$ cd ./some/working/directory
$ graven init
$ vi project.yaml

# Typical development cycle
$ graven clean
$ graven build
$ graven test
$ graven package

# When you're ready to release
$ graven release --login
$ graven release
$ graven bump [major|minor|patch|QUALIFIER]
```

A typical development cycle looks like this. The `init` command is run once per project, 
then clean, build, test and package are typically used throughout the development cycle.
Releases occur less frequently, and versions are bumped after the release.

The `freeze` and `unfreeze` commands are on a completely independent flow and can be 
executed any time. 

```
                               +----------+                           
                             > |  build   | \                         
                            /  +----------+  \                        
                           /                  \                       
                          /                    v                      
 +----------+    +----------+                 +----------+            
 |   init   |--->|   clean  |                 |   test   |            
 +----------+    +----------+                 +----------+            
                           ^                   /                        
                            \                 /                         
                             \ +----------+  /                          
                              \| package  |<-                           
    +----------+               +----------+                           
    |  freeze  |                     |                                
    +----------+                     v                                
          |                    +----------+                           
          |                    |  release |                           
    +-----v----+               +----------+                           
    | unfreeze |                     |                                
    +----------+                     |                                
                               +-----v----+                           
                               |   bump   |                           
                               +----------+                           
```


