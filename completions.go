package openai

import (
	"context"
	"encoding/json"

	"github.com/fabiustech/openai/models"
	"github.com/fabiustech/openai/objects"
	"github.com/fabiustech/openai/routes"
)

// CompletionRequest contains all relevant fields for requests to the completions endpoint.
type CompletionRequest[T models.Completion | models.FineTunedModel] struct {
	// Model specifies the ID of the model to use.
	// See more here: https://beta.openai.com/docs/models/overview
	Model T `json:"model"`
	// Prompt specifies the prompt(s) to generate completions for, encoded as a string, array of strings, array of
	// tokens, or array of token arrays. Note that <|endoftext|> is the document separator that the model sees during
	// training, so if a prompt is not specified the model will generate as if from the beginning of a new document.
	// Defaults to <|endoftext|>.
	Prompt string `json:"prompt,omitempty"`
	// Suffix specifies the suffix that comes after a completion of inserted text.
	// Defaults to null.
	Suffix string `json:"suffix,omitempty"`
	// MaxTokens specifies the maximum number of tokens to generate in the completion. The token count of your prompt
	// plus max_tokens cannot exceed the model's context length. Most models have a context length of 2048 tokens
	// (except for the newest models, which support 4096).
	// Defaults to 16.
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature specifies what sampling temperature to use. Higher values means the model will take more risks. Try
	// 0.9 for more creative applications, and 0 (argmax sampling) for ones with a well-defined answer. OpenAI generally
	// recommends altering this or top_p but not both.
	//
	// More on sampling temperature: https://towardsdatascience.com/how-to-sample-from-language-models-682bceb97277
	//
	// Defaults to 1.
	Temperature *float64 `json:"temperature,omitempty"`
	// TopP specifies an alternative to sampling with temperature, called nucleus sampling, where the model considers
	// the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10%
	// probability mass are considered. OpenAI generally recommends altering this or temperature but not both.
	// Defaults to 1.
	TopP *float64 `json:"top_p,omitempty"`
	// N specifies how many completions to generate for each prompt.
	// Note: Because this parameter generates many completions, it can quickly consume your token quota. Use carefully
	// and ensure that you have reasonable settings for max_tokens and stop.
	// Defaults to 1.
	N int `json:"n,omitempty"`
	// Steam specifies Whether to stream back partial progress. If set, tokens will be sent as data-only server-sent
	// events as they become available, with the stream terminated by a data: [DONE] message.
	// Defaults to false.
	Stream bool `json:"stream,omitempty"`
	// LogProbs specifies to include the log probabilities on the logprobs most likely tokens, as well the chosen
	// tokens. For example, if logprobs is 5, the API will return a list of the 5 most likely tokens. The API will
	// always return the logprob of the sampled token, so there may be up to logprobs+1 elements in the response.
	// The maximum value for logprobs is 5.
	// Defaults to null.
	LogProbs *int `json:"logprobs,omitempty"`
	// Echo specifies to echo back the prompt in addition to the completion.
	// Defaults to false.
	Echo bool `json:"echo,omitempty"`
	// Stop specifies up to 4 sequences where the API will stop generating further tokens. The returned text will not
	// contain the stop sequence.
	Stop []string `json:"stop,omitempty"`
	// PresencePenalty can be a number between -2.0 and 2.0. Positive values penalize new tokens based on whether they
	// appear in the text so far, increasing the model's likelihood to talk about new topics.
	// Defaults to 0.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`
	// FrequencyPenalty can be a number between -2.0 and 2.0. Positive values penalize new tokens based on their
	// existing frequency in the text so far, decreasing the model's likelihood to repeat the same line verbatim.
	// Defaults to 0.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	// Generates best_of completions server-side and returns the "best" (the one with the highest log probability per
	// token). Results cannot be streamed. When used with n, best_of controls the number of candidate completions and n
	// specifies how many to return – best_of must be greater than n. Note: Because this parameter generates many
	// completions, it can quickly consume your token quota. Use carefully and ensure that you have reasonable settings
	// for max_tokens and stop.
	// Defaults to 1.
	BestOf int `json:"best_of,omitempty"`
	// LogitBias modifies the likelihood of specified tokens appearing in the completion. Accepts a json object that
	// maps tokens (specified by their token ID in the GPT tokenizer) to an associated bias value from -100 to 100.
	// Mathematically, the bias is added to the logits generated by the model prior to sampling. The exact effect will
	// vary per model, but values between -1 and 1 should decrease or increase likelihood of selection; values like
	// -100 or 100 should result in a ban or exclusive selection of the relevant token.
	// As an example, you can pass {"50256": -100} to prevent the <|endoftext|> token from being generated.
	//
	// You can use this tokenizer tool to convert text to token IDs:
	// https://beta.openai.com/tokenizer
	//
	// Defaults to null.
	LogitBias map[string]int `json:"logit_bias,omitempty"`
	// User is a unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse.
	// See more here: https://beta.openai.com/docs/guides/safety-best-practices/end-user-ids
	User string `json:"user,omitempty"`
}

// CompletionChoice represents one of possible completions.
type CompletionChoice struct {
	Text         string         `json:"text"`
	Index        int            `json:"index"`
	FinishReason string         `json:"finish_reason"`
	LogProbs     *LogprobResult `json:"logprobs"`
}

// LogprobResult represents logprob result of Choice.
type LogprobResult struct {
	Tokens        []string             `json:"tokens"`
	TokenLogprobs []float32            `json:"token_logprobs"`
	TopLogprobs   []map[string]float32 `json:"top_logprobs"`
	TextOffset    []int                `json:"text_offset"`
}

// CompletionResponse is the response from the completions endpoint.
type CompletionResponse[T models.Completion | models.FineTunedModel] struct {
	ID      string              `json:"id"`
	Object  objects.Object      `json:"object"`
	Created uint64              `json:"created"`
	Model   T                   `json:"model"`
	Choices []*CompletionChoice `json:"choices"`
	Usage   *Usage              `json:"usage"`
}

// CreateCompletion creates a completion for the provided prompt and parameters.
func (c *Client) CreateCompletion(ctx context.Context, cr *CompletionRequest[models.Completion]) (*CompletionResponse[models.Completion], error) {
	var b, err = c.post(ctx, routes.Completions, cr)
	if err != nil {
		return nil, err
	}

	var resp = &CompletionResponse[models.Completion]{}
	if err = json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateFineTunedCompletion creates a completion for the provided prompt and parameters, using a fine-tuned model.
func (c *Client) CreateFineTunedCompletion(ctx context.Context, cr *CompletionRequest[models.FineTunedModel]) (*CompletionResponse[models.FineTunedModel], error) {
	var b, err = c.post(ctx, routes.Completions, cr)
	if err != nil {
		return nil, err
	}

	var resp = &CompletionResponse[models.FineTunedModel]{}
	if err = json.Unmarshal(b, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
