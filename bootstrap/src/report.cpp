#include "report.hpp"

#include <iostream>
#include <fstream>
#include <vector>
#include <limits>

#include "boost/algorithm/string/replace.hpp"

namespace balgo = boost::algorithm;

namespace chai {
    // appendRepeated adds `c` repeated `n` times to `str.
    static void appendRepeated(std::string& str, char c, int n) {
        for (int i = 0; i < n; i++)
            str.push_back(c);
    }

    // displayCompileMessage displays a compile message (error or warning).
    static void displayCompileMessage(bool isError, const CompileError& cerr) {
        // Display the error message header.
        std::cout << (isError ? "[error] " : "[warning] ")
            << cerr.displayPath() 
            << ':' << cerr.span().startLine
            << ':' << cerr.span().startCol
            << ": " << cerr.message()
            << "\n\n";

        // Find all the lines over which the error occurs.
        std::vector<std::string> lines;

        std::ifstream inf { cerr.fileAbsPath().data() };
        if (!inf) {
            reportICE("could not open file for error reporting\n");
        }

        for (int n = 0; inf; n++) {
            std::string input;
            std::getline(inf, input);

            if (cerr.span().startLine <= n && n <= cerr.span().endLine) {
                balgo::replace_all(input, "\t", "    ");
                lines.push_back(input);
            }
        }

        // Calculate the minimum indentation of the lines.
        int minIndent = std::numeric_limits<int>::max();
        for (auto& line : lines) {
            int lineIndent = 0;
            for (char c : line) {
                if (c == ' ')
                    lineIndent++;
                else if (c == '\n') {
                    lineIndent = -1;
                    break;
                } else 
                    break;
            }

            if (lineIndent == -1)
                continue;

            if (lineIndent < minIndent)
                minIndent = lineIndent;
        }

        // Calculate the maximum length of the line numbers when converted into
        // strings.  Since the end line number always has the greatest value and
        // thus length, it is sufficient to use its length for this metric.
        auto maxLineNumLen = std::to_string(cerr.span().endLine).length();

        int i = 0;
        for (auto& line : lines) {
            // Print the line number and padding bar.
            std::string lineNumber = std::to_string(cerr.span().startLine + i);
            std::cout << lineNumber;

            std::string buff;
            appendRepeated(buff, ' ', maxLineNumLen - lineNumber.length());

            std::cout << buff << " | ";

            // Print the line with the minimum indent trimmed off.
            std::cout << line.substr(minIndent) << '\n';

            // Print the padding bar for underlining line.
            buff.clear();
            appendRepeated(buff, ' ', maxLineNumLen + 1);
            std::cout << buff << "| ";

            // Calculate the number of spaces before carret underlining begins.
            // For any line which is not the starting line, this is always zero
            // since the underlining is always continuing form the previous
            // line.  For all other lines, it is start column - the minimum
            // indent.
            int carretPrefixCount = 0;
            if (i == 0)
                carretPrefixCount = cerr.span().startCol - minIndent;

            // Calculate the number of characters at the end of the source line
            // that should not be highlighted.  For all lines except the last
            // line, this is zero, since underlining should span until the end
            // of the line and over onto the next line.  For the last line, it
            // is length of the line - the end column of the errorenous source
            // text.
            int carretSuffixCount = 0;
            if (i == lines.size() - 1)
                carretSuffixCount = line.length() - cerr.span().endCol;

            // Print the number of spaces that come before the carret (ie. skip
            // underlining until the start column).
            buff.clear();
            appendRepeated(buff, ' ', carretPrefixCount);
            
            std::cout << buff;
            
            // Print the underlining carrets for the given line.  This number is
            // always the length of the line - (the number of carret that are
            // skipped at the start + the number of carrets that are skipped at
            // the end).  We also subtract off the minimum indentation to
            // account for the fact that that part of the line is neglected by
            // the prefix count (ie. to cancel the - minimum indent inside the
            // calculate for carret prefix count).
            buff.clear();
            appendRepeated(buff, '^', line.length() - carretPrefixCount - carretSuffixCount - minIndent);

            std::cout << buff << '\n';

            i++;
        }
    }

    // -------------------------------------------------------------------------- //
    
    LogLevel reporterLogLevel = LogLevel::VERBOSE;

    void reportError(const CompileError& cerr) {
        if (reporterLogLevel > LogLevel::SILENT) {
            displayCompileMessage(true, cerr);
        }
    }

    void reportError(std::string_view displayPath, const std::exception& e) {
        if (reporterLogLevel > LogLevel::SILENT) {
            std::cout << "[error] " << displayPath << ": " << e.what() << '\n';
        }
    }

    void reportWarning(const CompileError& cerr) {
        if (reporterLogLevel > LogLevel::SILENT) {
            displayCompileMessage(false, cerr);
        }
    }

    void internalReportFatal(const std::string& message) {
        if (reporterLogLevel > LogLevel::SILENT) {
            std::cout << "[error] " << message << '\n';
        }

        exit(1);
    }

    void reportFatal(const std::exception& e) {
        if (reporterLogLevel > LogLevel::SILENT) {
            std::cout << "[error] " << e.what() << '\n';
        }

        exit(1);
    }

    void internalReportICE(const std::string& message) {
        std::cout << "[internal compiler error] " << message << '\n';
        std::cout << "This was not supposed to happen.  Please open an issue on GitHub.\n";

        exit(-1);
    }

    void reportICE(const std::exception& e) {
        std::cout << "[internal compiler error] " << e.what() << '\n';
        std::cout << "This was not supposed to happen.  Please open an issue on GitHub.\n";

        exit(-1);
    }
}