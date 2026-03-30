// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

package helixqa

import (
	"digital.vasic.llmsverifier/pkg/strategy"
)

// VisionModelRegistry returns all known vision-capable models for HelixQA
// verification. These entries feed into the QAStrategy scoring and ranking
// pipeline so that new providers are included in the verification matrix.
func VisionModelRegistry() []strategy.ModelInfo {
	return []strategy.ModelInfo{
		// --- Tier 1: Premium vision providers ---
		{
			ID:                      "openai-gpt4o",
			Name:                    "GPT-4o",
			Provider:                "openai",
			Model:                   "gpt-4o",
			SupportsVision:          true,
			SupportsStreaming:       true,
			SupportsFunctionCalling: true,
			ContextWindow:           128000,
			MaxOutputTokens:         4096,
			AvgLatencyMs:            1200,
			InputCostPer1k:          0.005,
			OutputCostPer1k:         0.015,
			QualityScore:            0.95,
			ReliabilityScore:        0.98,
			Capabilities:            []string{"vision", "code", "reasoning", "gui_analysis"},
		},
		{
			ID:                      "anthropic-sonnet",
			Name:                    "Claude Sonnet 4",
			Provider:                "anthropic",
			Model:                   "claude-sonnet-4-20250514",
			SupportsVision:          true,
			SupportsStreaming:       true,
			SupportsFunctionCalling: true,
			ContextWindow:           200000,
			MaxOutputTokens:         4096,
			AvgLatencyMs:            1500,
			InputCostPer1k:          0.003,
			OutputCostPer1k:         0.015,
			QualityScore:            0.94,
			ReliabilityScore:        0.97,
			Capabilities:            []string{"vision", "code", "reasoning", "gui_analysis"},
		},
		{
			ID:                      "gemini-flash",
			Name:                    "Gemini 2.0 Flash",
			Provider:                "google",
			Model:                   "gemini-2.0-flash",
			SupportsVision:          true,
			SupportsStreaming:       true,
			SupportsFunctionCalling: true,
			ContextWindow:           1048576,
			MaxOutputTokens:         8192,
			AvgLatencyMs:            800,
			InputCostPer1k:          0.0001,
			OutputCostPer1k:         0.0004,
			QualityScore:            0.88,
			ReliabilityScore:        0.96,
			Capabilities:            []string{"vision", "code", "reasoning", "gui_analysis"},
		},

		// --- Tier 2: Cost-effective vision providers ---
		{
			ID:                      "kimi-k25",
			Name:                    "Kimi K2.5",
			Provider:                "kimi",
			Model:                   "kimi-k2.5",
			SupportsVision:          true,
			SupportsStreaming:       true,
			SupportsFunctionCalling: true,
			ContextWindow:           131072,
			MaxOutputTokens:         8192,
			AvgLatencyMs:            1000,
			InputCostPer1k:          0.0003,
			OutputCostPer1k:         0.0006,
			QualityScore:            0.85,
			ReliabilityScore:        0.90,
			Capabilities:            []string{"vision", "code", "reasoning", "gui_analysis"},
			Metadata: map[string]any{
				"architecture": "1T MoE",
				"license":      "MIT",
				"api_compat":   "openai",
			},
		},
		{
			ID:                "stepgui-v15",
			Name:              "Step-GUI 1.5V Mini",
			Provider:          "stepfun",
			Model:             "step-1.5v-mini",
			SupportsVision:    true,
			SupportsStreaming:  true,
			ContextWindow:     32768,
			MaxOutputTokens:   8192,
			AvgLatencyMs:      900,
			InputCostPer1k:    0.0,
			OutputCostPer1k:   0.0,
			QualityScore:      0.82,
			ReliabilityScore:  0.85,
			Capabilities:      []string{"vision", "gui_grounding", "gui_navigation", "element_detection"},
			Metadata: map[string]any{
				"specialization": "GUI interaction",
				"free_tier":      true,
			},
		},
		{
			ID:                "qwen3-vl",
			Name:              "Qwen3-VL",
			Provider:          "qwen",
			Model:             "qwen-vl-max",
			SupportsVision:    true,
			SupportsStreaming:  true,
			ContextWindow:     32768,
			MaxOutputTokens:   4096,
			AvgLatencyMs:      1100,
			InputCostPer1k:    0.001,
			OutputCostPer1k:   0.002,
			QualityScore:      0.87,
			ReliabilityScore:  0.88,
			Capabilities:      []string{"vision", "gui_grounding", "code", "reasoning"},
			Metadata: map[string]any{
				"ui_grounding_accuracy": 0.90,
				"open_source":           true,
			},
		},

		// --- Tier 3: Local/open-source vision providers ---
		{
			ID:               "ollama-llava",
			Name:             "LLaVA 7B (Ollama)",
			Provider:         "ollama",
			Model:            "llava:7b",
			SupportsVision:   true,
			SupportsStreaming: true,
			ContextWindow:    4096,
			MaxOutputTokens:  2048,
			AvgLatencyMs:     3000,
			InputCostPer1k:   0.0,
			OutputCostPer1k:  0.0,
			QualityScore:     0.65,
			ReliabilityScore: 0.80,
			Capabilities:     []string{"vision", "local"},
			Metadata: map[string]any{
				"deployment": "local",
				"free":       true,
			},
		},
	}
}

// BudgetVisionModels returns only cost-effective vision models
// ($0 or < $1/1M tokens) suitable for high-volume QA testing.
func BudgetVisionModels() []strategy.ModelInfo {
	var budget []strategy.ModelInfo
	for _, m := range VisionModelRegistry() {
		totalCost := m.InputCostPer1k + m.OutputCostPer1k
		if totalCost < 0.002 { // < $2/1k = < $2/M tokens
			budget = append(budget, m)
		}
	}
	return budget
}

// GUISpecializedModels returns models specialized for GUI interaction
// and UI element grounding.
func GUISpecializedModels() []strategy.ModelInfo {
	var gui []strategy.ModelInfo
	for _, m := range VisionModelRegistry() {
		for _, cap := range m.Capabilities {
			if cap == "gui_grounding" || cap == "gui_navigation" {
				gui = append(gui, m)
				break
			}
		}
	}
	return gui
}
