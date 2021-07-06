package build

import (
	"chai/logging"
	"chai/mods"
	"chai/sem"
	"chai/syntax"
	"fmt"
)

// processMetadata scans the top of a file for metadata, validates it, adds it
// to the file as necessary, and indicates whether or not the compiler should
// compile the file as necessary.  It takes the scanner for the file as its
// second argument (so that it can read metadata easily)
func (c *Compiler) processMetadata(file *sem.ChaiFile, sc *syntax.Scanner) bool {
	// scan metadata
	if metadata, ok := c.parseMetadata(sc); ok {
		// check for compilation conditions
		if _, ok := metadata["no_compile"]; ok {
			return false
		}

		if osName, ok := metadata["os"]; ok {
			if _, ok := mods.OSNames[osName]; ok && osName == c.buildProfile.TargetOS {
				return true
			}

			return false
		}

		if archName, ok := metadata["arch"]; ok {
			if _, ok := mods.ArchNames[archName]; ok && archName == c.buildProfile.TargetArch {
				return true
			}

			return false
		}

		file.Metadata = metadata
		return true
	}

	return false
}

// parseMetadata uses a file's scanner to extract any metadata at the top of the
// file.  It returns a map of metadata arguments to be processed by the compiler
// and a boolean indicating whether or not metadata was scanned successfully.
func (c *Compiler) parseMetadata(sc *syntax.Scanner) (map[string]string, bool) {
	logUnexpectedTokenError := func(tok *syntax.Token) {
		logging.LogCompileError(
			sc.Context(),
			fmt.Sprintf("unexpected token: `%s`", tok.Value),
			logging.LMKSyntax,
			syntax.TextPositionOfToken(tok),
		)
	}

	// check to see if the file starts with two exclamation marks (denoting
	// metadata) -- unread them if it doesn't (so we don't mess with the
	// scanner)
	if firstTok, ok := sc.ReadToken(); ok {
		if firstTok.Kind == syntax.NOT {
			if nextTok, ok := sc.ReadToken(); ok {
				if nextTok.Kind == syntax.NOT {
					// closer is a newline in the normal case; a closing paren
					// in the case of multiline metadata
					closer := syntax.NEWLINE
					allowEnd := true
					metadata := make(map[string]string)
					state := 0
					recentKey := ""

					for {
						nextTok, ok = sc.ReadToken()
						if !ok {
							return metadata, allowEnd
						}

						switch state {
						// initial state
						case 0:
							switch nextTok.Kind {
							case syntax.LPAREN:
								closer = syntax.RPAREN
								allowEnd = false
								state = 1
							case syntax.IDENTIFIER:
								recentKey = nextTok.Value

								if _, ok := metadata[recentKey]; ok {
									logging.LogCompileWarning(
										sc.Context(),
										fmt.Sprintf("metadata parameter `%s` specified multiple times", recentKey),
										logging.LMKMetadata,
										syntax.TextPositionOfToken(nextTok),
									)
								}

								state = 2
							case closer:
								return metadata, true
							}
						// expecting identifier
						case 1:
							if nextTok.Kind == syntax.IDENTIFIER {
								recentKey = nextTok.Value

								if _, ok := metadata[recentKey]; ok {
									logging.LogCompileWarning(
										sc.Context(),
										fmt.Sprintf("metadata parameter `%s` specified multiple times", recentKey),
										logging.LMKMetadata,
										syntax.TextPositionOfToken(nextTok),
									)
								}

								state = 2
							} else {
								logUnexpectedTokenError(nextTok)
								return nil, false
							}
						// expecting equals or comma
						case 2:
							if nextTok.Kind == syntax.ASSIGN {
								state = 3
							} else if nextTok.Kind == syntax.COMMA {
								state = 1
							} else {
								logUnexpectedTokenError(nextTok)
								return nil, false
							}
						// expecting value
						case 3:
							if nextTok.Kind == syntax.STRING {
								metadata[recentKey] = nextTok.Value
								state = 4
							} else {
								logUnexpectedTokenError(nextTok)
								return nil, false
							}
						case 4:
							if nextTok.Kind == syntax.COMMA {
								state = 1
							} else if nextTok.Kind == closer {
								return metadata, true
							} else {
								logUnexpectedTokenError(nextTok)
								return nil, false
							}
						}
					}

				} else {
					// a file that starts with an exclamation but does not use
					// metadata is unparsable
					logUnexpectedTokenError(nextTok)
					return nil, false
				}
			}
		} else {
			sc.UnreadToken(firstTok)
			return make(map[string]string), true
		}
	}

	// shouldn't compile because the file is empty :)
	return nil, false
}
