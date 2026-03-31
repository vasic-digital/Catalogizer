// Package challenges provides the challenge bank definitions, registration, configuration,
// and execution logic for the Catalogizer quality assurance system.
//
// Challenges are structured test scenarios that validate end-to-end functionality of the
// Catalogizer system. They are registered via RegisterAll and exposed through /api/v1/challenges
// REST endpoints. This package includes both the original challenge bank (CH-001 to CH-050),
// user flow challenges (UF-*), and module verification challenges (MOD-*).
package challenges
