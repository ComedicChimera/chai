#ifndef _REPORT_H_
#define _REPORT_H_

#include <stdexcept>

#include "chaic.hpp"

namespace chai {
    // LogLevel enumerates the logging output levels of the compiler.
    enum class LogLevel {
        SILENT,   // No output.
        ERROR,    // Only errors.
        WARN,     // Only errors and warnings.
        VERBOSE,  // All output.
    };

    // CompileError represents an error that occurred during compilation.
    class CompileError : public std::exception {
        std::string_view m_displayPath;
        std::string_view m_fileAbsPath;
        std::string m_message;
        TextSpan m_span;

    public:
        // Creates a new compile error occuring in the file at `fileAbsPath`,
        // with error message `message`, and text span `span`.
        CompileError(
            std::string_view displayPath, 
            std::string_view fileAbsPath, 
            std::string&& message, 
            const TextSpan& span
        ) 
        : m_displayPath(displayPath)
        , m_fileAbsPath(fileAbsPath)
        , m_message(message)
        , m_span(span)
        {}

        // displayPath returns a view to the display path to the source file the
        // compile error occurs in.  The display path is used to concisely
        // identify the erroneous file to the user during error reporting.
        inline std::string_view displayPath() const { return m_displayPath; }

        // fileAbsPath returns a view to the absolute path to the source file
        // the compile error occurs in.
        inline std::string_view fileAbsPath() const { return m_fileAbsPath; }

        // message returns a view to the compile error's error message.
        inline std::string_view message() const { return m_message; }

        // span returns the span where the compile error occurs in source text.
        inline const TextSpan& span() const { return m_span; }
    };

    // The global log level.
    extern LogLevel reporterLogLevel;

    // reportError reports a compile error.
    void reportError(const CompileError& cerr);

    // reportError reports a standard exception.
    void reportError(std::string_view displayPath, const std::exception& e);

    // reportWarning reports a compile warning.
    void reportWarning(const CompileError& cerr);

    // internalReportFatal is an internal implementation of reportFatal that
    // does not require templates so as to avoid inlining the full definition of
    // reportFatal in the header.
    void internalReportFatal(const std::string& message);

    // reportFatal reports a fatal error message.
    template<typename ...T>
    inline void reportFatal(const std::string& fmt, const T& ...args) {
        internalReportFatal(std::format(fmt, args...));
    }

    // reportFatal reports a standard exception as a fatal error.
    void reportFatal(const std::exception& e);

    // internalReportICE is an internal implementation of reportICE that does
    // not require templates so as to avoid inlining the full definition of
    // reportICE in the header.
    void internalReportICE(const std::string& message);

    // reportICE reports a internal compiler error message.
    template<typename ...T>
    inline void reportICE(const std::string& fmt, const T& ...args) {
        internalReportICE(std::format(fmt, args...));
    }

    // reportICE reports an exception as internal compiler error
    void reportICE(const std::exception& e);
}

#endif