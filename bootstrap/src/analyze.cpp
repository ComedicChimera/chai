#include "compiler.hpp"

#include <cctype>
#include <fstream>
#include <iostream>

#include "report.hpp"
#include "syntax/lexer.hpp"

namespace chai {
    // isValidIdentifier returns if the string is a valid identifier.
    static bool isValidIdentifier(std::string_view identName) {
        if (identName.length() == 0)
            return false;

        if (isalpha(identName[0]) || identName[0] == '_') {
            for (int i = 1; i < identName.length(); i++) {
                if (!isalnum(identName[i]) && identName[i] != '_')
                    return false;
            }

            return true;
        }
        
        return false;
    }

    /* ---------------------------------------------------------------------- */

    ChaiPackage* Compiler::initPackage(const std::filesystem::path& absPath) {
        // Create the package and add it to the dependency graph.
        auto* pkg = new ChaiPackage(absPath);
        m_depGraph.insert(std::make_pair(pkg->id, std::unique_ptr<ChaiPackage>(pkg)));

        // Validate the package name.
        if (!isValidIdentifier(pkg->name())) 
            reportFatal("{} is not a valid package name", pkg->name());

        // Walk all the source files in the package directory.
        for (const auto& file : std::filesystem::directory_iterator(pkg->absPath())) {
            // We only want to try to load source files.
            if (file.is_directory() || file.path().extension() != ".chai")
                continue;

            // Create the Chai source file and add it to its parent package.
            auto* chFile = new ChaiFile(pkg, pkg->files.size(), std::filesystem::absolute(file.path()));
            pkg->files.push_back(std::unique_ptr<ChaiFile>(chFile));  

            // DEBUG
            std::ifstream inf (chFile->absPath());
            Lexer lexer(inf, chFile);      

            try {
                Token token;
                while ((token = lexer.nextToken()).kind != TokenKind::END_OF_FILE) {
                    std::cout << (int)(token.kind) << ", " << token.value << "\n";
                }
            } catch (CompileError& cerr) {
                reportError(cerr);
            } catch (std::exception* err) {
                reportFatal(*err);
            }

            inf.close();
        }

        // Make sure the package is non-empty.
        if (pkg->files.size() == 0)
            reportFatal("package %s must contain source files", pkg->name());

        return pkg;
    }
}