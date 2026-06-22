#!/usr/bin/env bash
# =============================================================================
# test_constitution_inheritance.sh — host-side test for the constitution-
#   inheritance gate + its paired §1.1 meta-test
# -----------------------------------------------------------------------------
# Purpose:
#   Drive BOTH halves of the constitution-inheritance verification and print a
#   clear PASS/FAIL summary for each:
#     1. The GATE — scripts/verify_constitution_inheritance.sh — asserts the
#        submodule is present + inherited (INV1..INV5).
#     2. The META-TEST — scripts/meta_test_constitution_inheritance.sh —
#        proves the gate has teeth (Constitution §1.1 paired mutation: mutate
#        Constitution.md, run the gate, assert it FAILs, restore).
#   Exits non-zero if EITHER fails, so a CI / pre-build lane breaks on a missing
#   or non-inherited constitution OR on a bluff gate.
#
# Usage:
#   bash tests/test_constitution_inheritance.sh
#     (run from the project root: /Volumes/T7/Projects/catalogizer)
#
# Inputs:
#   - $PROJECT_ROOT (optional) — overrides the auto-resolved project root.
#
# Outputs:
#   - Per-check captured output + a PASS/FAIL summary line for the GATE and the
#     META-TEST, plus an overall verdict, on stdout/stderr.
#   - Exit 0 ONLY if both the GATE and the META-TEST pass; non-zero otherwise.
#
# Side-effects:
#   - The GATE is read-only. The META-TEST drives the constitution-side helper
#     which temporarily mutates + restores Constitution.md (snapshot/restore).
#
# Dependencies:
#   - bash, scripts/verify_constitution_inheritance.sh,
#     scripts/meta_test_constitution_inheritance.sh, the constitution submodule.
#
# Cross-references:
#   - Constitution §1.1 / §11.4 / §11.4.18 / §11.4.35.
#   - scripts/verify_constitution_inheritance.sh (the gate).
#   - scripts/meta_test_constitution_inheritance.sh (the paired meta-test).
# =============================================================================

set -uo pipefail   # NOT -e: we want to run BOTH checks and report both verdicts.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$(cd "${SCRIPT_DIR}/.." && pwd)}"
cd "${PROJECT_ROOT}"

GATE="scripts/verify_constitution_inheritance.sh"
META="scripts/meta_test_constitution_inheritance.sh"

GATE_RC=0
META_RC=0

echo "############################################################"
echo "# Constitution-inheritance host-side test"
echo "# project root: ${PROJECT_ROOT}"
echo "############################################################"
echo
echo "------------------------------------------------------------"
echo "[1/2] GATE: ${GATE}"
echo "------------------------------------------------------------"
if [[ -f "${PROJECT_ROOT}/${GATE}" ]]; then
    bash "${PROJECT_ROOT}/${GATE}"
    GATE_RC=$?
else
    echo "FAIL: gate script missing at ${GATE}" >&2
    GATE_RC=2
fi
echo

echo "------------------------------------------------------------"
echo "[2/2] META-TEST: ${META}"
echo "------------------------------------------------------------"
if [[ -f "${PROJECT_ROOT}/${META}" ]]; then
    bash "${PROJECT_ROOT}/${META}"
    META_RC=$?
else
    echo "FAIL: meta-test script missing at ${META}" >&2
    META_RC=2
fi
echo

# --- Summary -----------------------------------------------------------------
echo "############################################################"
echo "# SUMMARY"
echo "############################################################"
if [[ "${GATE_RC}" -eq 0 ]]; then
    echo "  GATE      : PASS"
else
    echo "  GATE      : FAIL (rc=${GATE_RC})"
fi
if [[ "${META_RC}" -eq 0 ]]; then
    echo "  META-TEST : PASS"
else
    echo "  META-TEST : FAIL (rc=${META_RC})"
fi

if [[ "${GATE_RC}" -eq 0 && "${META_RC}" -eq 0 ]]; then
    echo "  OVERALL   : PASS"
    exit 0
else
    echo "  OVERALL   : FAIL" >&2
    exit 1
fi
