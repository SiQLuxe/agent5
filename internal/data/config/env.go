package config

import "os"

func applyEnvOverrides(cfg *Config) {
    if val := os.Getenv("OPENAI_API_KEY"); val != "" {
        cfg.Models.OpenAI.APIKey = val
    }
    if val := os.Getenv("OPENAI_BASE_URL"); val != "" {
        cfg.Models.OpenAI.BaseURL = val
    }
    if val := os.Getenv("OPENAI_DEFAULT_MODEL"); val != "" {
        cfg.Models.OpenAI.DefaultModel = val
    }

    if val := os.Getenv("DEEPSEEK_API_KEY"); val != "" {
        cfg.Models.DeepSeek.APIKey = val
    }
    if val := os.Getenv("DEEPSEEK_BASE_URL"); val != "" {
        cfg.Models.DeepSeek.BaseURL = val
    }
    if val := os.Getenv("DEEPSEEK_DEFAULT_MODEL"); val != "" {
        cfg.Models.DeepSeek.DefaultModel = val
    }

    if val := os.Getenv("LOCAL_BASE_URL"); val != "" {
        cfg.Models.Local.BaseURL = val
    }
    if val := os.Getenv("LOCAL_DEFAULT_MODEL"); val != "" {
        cfg.Models.Local.DefaultModel = val
    }

    if val := os.Getenv("DEFAULT_CLIENT"); val != "" {
        cfg.DefaultClient = val
    }
}
