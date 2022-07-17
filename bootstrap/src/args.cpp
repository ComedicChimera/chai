#include "compiler.hpp"

#include <unordered_map>
#include <iostream>
#include <filesystem>

#include "boost/program_options.hpp"

#include "chaic.hpp"
#include "report.hpp"

namespace po = boost::program_options;

namespace chai {
    // OptionKind represents a command-line option kind: it is used to map
    // option names to integral values so they can be switched over.
    enum class OptionKind {
        ROOTDIR,  // <root directory>
        HELP,     // --help
        VERSION,  // --version
        DEBUG,    // --debug
        OUTPATH,  // --outpath
        OUTMODE,  // --outmode
        LOGLEVEL, // --loglevel
    };

    // A table mapping option names to integral option kinds.
    std::unordered_map<std::string, OptionKind> optionNameToKind {
        { "rootdir", OptionKind::ROOTDIR },
        { "help", OptionKind::HELP },
        { "version", OptionKind::VERSION },
        { "debug", OptionKind::DEBUG },
        { "outpath", OptionKind::OUTPATH },
        { "outmode", OptionKind::OUTMODE },
        { "loglevel", OptionKind::LOGLEVEL },
    };

    // A table mapping output mode names to OutputMode values.
    std::unordered_map<std::string, OutputMode> modeNameToEnum {
        { "exe", OutputMode::EXE },
        { "obj", OutputMode::OBJ },
        { "asm", OutputMode::ASM },
        { "llvm", OutputMode::LLVM },
    };

    // A table mapping log level names to LogLevel values.
    std::unordered_map<std::string, LogLevel> logLevelNameToEnum {
        { "silent", LogLevel::SILENT },
        { "error", LogLevel::ERROR },
        { "warn", LogLevel::WARN },
        { "verbose", LogLevel::VERBOSE },
    };

    bool Compiler::initFromArgs(int argc, char* argv[]) {
        // Build the command-line parser.
        po::options_description visibleDesc("usage: chaic [options] <root directory>\n\nOptions");

        // TODO Should we move this code elsewhere so we don't have a giant wall
        // of text sitting in the middle of this function?
        visibleDesc.add_options()
            ("help,h", "Display usage information (ie. this text).")
            ("version,v", "Displays the current compiler version.")
            ("debug,d", "Whether the compiler should output debug information.")
            ("outpath,o", po::value<std::string>()->default_value("out")->value_name("<output path>"), 
                "Sets the path for compilation output.  "
                "If a single output is to be produced, this should be a file.  "
                "If multiple outputs are to be produced, this should be a directory.  "
                "Default to out[.<platform-extension>] if unspecified."
            )
            ("outmode,m", po::value<std::string>()->default_value("exe")->value_name("<mode>"), 
                "Sets the compilation mode: the type of output the compiler should produce.  "
                "Valid values are:\n"
                "  - \"exe\" for producing an executable (default)\n"
                "  - \"obj\" for producing object files for each package\n"
                "  - \"asm\" for producing assembly files for each package\n"
                "  - \"llvm\" for producing LLVM IR source files for each package\n"
            )
            ("loglevel", po::value<std::string>()->default_value("verbose")->value_name("<log level>"),
                "Sets the compiler's log level.  Valid values are:"
                "  - \"verbose\" for outputting all messages (default)\n"
                "  - \"warn\" for outputing errors and warnings\n"
                "  - \"error\" for outputting errors only\n"
                "  - \"silent\" for no output"
            )
        ;

        po::options_description hiddenDesc;
        hiddenDesc.add_options()
            ("rootdir", po::value<std::string>(), "the root directory")
        ;

        po::positional_options_description positionalDesc;
        positionalDesc.add("rootdir", 1);

        po::options_description combinedDesc;
        combinedDesc.add(visibleDesc).add(hiddenDesc);

        // Run the parser on the command-line arguments.
        po::variables_map parseResult;
        try {
            po::store(
                po::command_line_parser(argc, argv)
                .options(combinedDesc)
                .positional(positionalDesc)
                .run(), 
                parseResult
            );
            po::notify(parseResult);
        } catch (po::unknown_option& uoe) {
            std::cout << "[error] unknown option: " << uoe.get_option_name() << '\n';
            return false; 
        }

        // Apply each of the arguments to the compiler.
        bool hasRootDir = false, hasOutPath = false;
        for (auto& option : parseResult) {
            auto kind = optionNameToKind[option.first];

            switch (kind) {
            case OptionKind::ROOTDIR:
                m_rootDir = std::filesystem::absolute(option.second.as<std::string>()).string(); 
                hasRootDir = true;
                break;
            case OptionKind::HELP:
                std::cout << visibleDesc << '\n';
                return false;
            case OptionKind::VERSION:
                std::cout << CHAI_NAME << " " << CHAI_VERSION << '\n';
                return false;
            case OptionKind::DEBUG:
                m_debug = true;
                break;
            case OptionKind::OUTPATH:
                m_outputPath = std::filesystem::absolute(option.second.as<std::string>()).string();
                hasOutPath = true;
                break;
            case OptionKind::OUTMODE:
            {
                std::string& mode = option.second.as<std::string>();

                if (modeNameToEnum.find(mode) == modeNameToEnum.end()) {
                    std::cout << "[error] invalid output mode\n";
                    return false;
                } else {
                    m_mode = modeNameToEnum[mode];
                }

                break;
            }  
            case OptionKind::LOGLEVEL:
                // TODO
                break;
            }
        }

        // Assert that the root directory has been specified.
        if (!hasRootDir) {
            std::cout << "[error] specify a root directory\n";
            return false;
        }

        // Set the default output path if it is unspecified.
        if (!hasOutPath) {
            m_outputPath = "out";

            #ifdef _WIN32
                if (m_mode == OutputMode::EXE)
                    m_outputPath += ".exe";
            #endif
        }

        // If we reach here then initialization was successful and we return true.
        return true;
    }
}