# Overview

Some LLVM toolchains, (e.g., Apple LLVM toolchain) doesn't come with useful LLVM tooling to work with bitcode like `bcanalyzer`, `opt`, and `exegesis`.

Conjunct is a `clang` wrapper that allows inserting different flows in the compilation process.

A typical use case of Conjunct goes like this:
- A build system runs a command like `clang -c hello1.c -o hello1.o && clang -c hello2.c -o hello2.o` to generate two object files
- Later, the same build system would run `clang hello1.o hello2.o -o hello` to link the object files in a final binary
- If you want to run intermediate steps in-between the two steps,
    - you would usually modify the build system to inject those commands.
    - This might be an issue since the build system might not allow you to inject intermediate steps easily.
    - Another case is that your LLVM toolchain is lacking, and you would want to mix-and-match with your own LLVM toolchains

This is where Conjunct comes along:
- Most build systems would easily allow you to set the `CC` environment variable to specify you compiler
    - You'd specify Conjunct in this case (i.e., `CC=conjunct`)
- Then, write a Conjunct config file specifying the intermediate steps you want to run (see the `Config File Specs` section below) and pass its location as a compiler flag `i.e., "--conjunct-config-path=my-conjunct-config.yaml"`
- What happens next is that the build system would run the same command in the original flow but with Conjunct this time (i.e., `conjunct -c hello1.c -o hello1.o && clang -c hello2.c -o hello2.o`)
    - Right here, conjunct would build the object files, and then run whatever tooling you want on the generated bitcode
- Finally, your build system would run `conjunct hello1.o hello2.o -o hello` to generate the final binary
    - Conjunct would call the original `clang` to do this linking step
- The flow from the build system's view is the same, but we were able to add whatever intermediate steps we want to the flow using Conjunct as a wrapper

# Usage

This is how you use Conjunct:

- Make a config file (see `Config File Specs` section below) that fits your intermediate steps
- Build Conjunct
    - `mage buildConjunct`
- Set Conjunct as your compiler
- Pass `--conjunct-config-path` as a compiler flag

Conjunct uses [Mage](https://github.com/magefile/mage) as a task runner with the following tasks. Either view the `./magefile.go` or run `mage -l` to view the targets. Most targets have multiple parameters. Use `mage -h TARGET_NAME` to view the parameters (or until [this issue](https://github.com/magefile/mage/issues/482) is resolved).

For easy reference, to build Conjunct, just run `mage buildConjunct`

# Conjunct Parameters

Conjunct by default accepts all the parameters you'd regularly pass to Clang. Conjunct has a few special parameters as well:
- `--conjunct-config-path=<CONFIG_FILE_PATH>`
    - Path to the Conjunct config file (see `Config File Specs` section below)
- `--conjunct-verbose`
    - Activate verbosity
    - Very useful for debugging Conjunct
- `--conjunct-dry-run`
    - Run the tool but don't actually run any intermediate steps
    - This basically is the same as running `clang`, but print all the logs of the intermediate steps without runnning any commands.
    - Very useful for debugging Conjunct
- `--conjunct-retain-temp-dir`
    - Retain the temporary directory where all the intermediate steps dump their contents
    - Very useful for debugging Conjunct

# Config File Specs

To run any intermediate steps, you need a config file that specifies what needs to run. This is supplied to Conjunct through the `--conjunct-config-path=<CONFIG_FILE_PATH>` parameter.

The config file should be a YAML file. The specs are in `./config/config.go:ConjunctConfig`. There's an example file in the demos here: `./testassets/ios/ConjunctDemo/conjunct-config.yaml` and `./testassets/android/ConjunctDemo/conjunct-config.yaml`

# Testing

You can run the unit tests with `mage runUnitTests`.

There's also tests for both Android and iOS projects, run each with:

        // Android
        mage -v BuildAndroidDemoWithConjunct \
            ${OPT_PATH} \
            ${CLANG_DIR_PATH} \
            true

        // iOS
        mage -v BuildIosDemoWithConjunct \
            ${OPT_PATH} \
            ${CLANG_DIR_PATH} \
            true

Where:
- `${CLANG_DIR_PATH}` is pointing to directory containing both `clang` and `clang++` binaries
- `${OPT_PATH}` is pointing to an `opt` binary
