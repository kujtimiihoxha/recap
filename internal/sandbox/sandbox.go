// Package sandbox provides a goja-based JavaScript execution environment
// with injected documents and LLM query support for the RLM agent pattern.
package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
)

type LLMFunc func(ctx context.Context, llmContext string, query string) (string, error)

type Document struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Option func(*config)

type config struct {
	documents []Document
	llmFn     LLMFunc
}

func WithDocuments(docs []Document) Option {
	return func(c *config) { c.documents = docs }
}

func WithLLM(fn LLMFunc) Option {
	return func(c *config) { c.llmFn = fn }
}

type Sandbox struct {
	mu     sync.Mutex
	loop   *eventloop.EventLoop
	stdout *strings.Builder
	cfg    config
}

func New(opts ...Option) (*Sandbox, error) {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}

	sb := &Sandbox{
		loop:   eventloop.NewEventLoop(),
		stdout: &strings.Builder{},
		cfg:    cfg,
	}
	sb.loop.Start()
	return sb, nil
}

func (s *Sandbox) setGlobalJSON(vm *goja.Runtime, name string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = vm.RunString(fmt.Sprintf("var %s = %s;", name, string(data)))
	return err
}

func (s *Sandbox) registerPrint(vm *goja.Runtime) {
	_ = vm.Set("print", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		s.stdout.WriteString(strings.Join(parts, " "))
		s.stdout.WriteString("\n")
		return goja.Undefined()
	})
}

func (s *Sandbox) injectGlobals(vm *goja.Runtime, ctx context.Context) error {
	s.registerPrint(vm)
	if s.cfg.documents != nil {
		if err := s.setGlobalJSON(vm, "documents", s.cfg.documents); err != nil {
			return fmt.Errorf("inject documents: %w", err)
		}
	}
if s.cfg.llmFn != nil {
		s.setLLMQuery(vm, ctx)
	}
	return nil
}

func (s *Sandbox) setLLMQuery(vm *goja.Runtime, ctx context.Context) {
	_ = vm.Set("llm_query", func(call goja.FunctionCall) goja.Value {
		llmContext := call.Argument(0).String()
		query := call.Argument(1).String()

		promise, resolve, reject := vm.NewPromise()

		go func() {
			result, err := s.cfg.llmFn(ctx, llmContext, query)
			s.loop.RunOnLoop(func(vm *goja.Runtime) {
				if err != nil {
					_ = reject(vm.NewGoError(err))
				} else {
					_ = resolve(vm.ToValue(result))
				}
			})
		}()

		return vm.ToValue(promise)
	})
}

func (s *Sandbox) Eval(ctx context.Context, code string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stdout.Reset()

	wrapped := fmt.Sprintf("(async () => {\n%s\n})()", code)

	type result struct {
		err error
	}
	done := make(chan result, 1)

	s.loop.RunOnLoop(func(vm *goja.Runtime) {
		if err := s.injectGlobals(vm, ctx); err != nil {
			done <- result{fmt.Errorf("sandbox eval: %w", err)}
			return
		}

		v, err := vm.RunString(wrapped)
		if err != nil {
			done <- result{fmt.Errorf("sandbox eval: %w", err)}
			return
		}

		p, ok := v.Export().(*goja.Promise)
		if !ok {
			done <- result{}
			return
		}

		var poll func(*goja.Runtime)
		poll = func(vm *goja.Runtime) {
			switch p.State() {
			case goja.PromiseStateFulfilled:
				done <- result{}
			case goja.PromiseStateRejected:
				done <- result{fmt.Errorf("sandbox eval: %s", p.Result().String())}
			default:
				s.loop.RunOnLoop(poll)
			}
		}
		s.loop.RunOnLoop(poll)
	})

	select {
	case <-ctx.Done():
		return s.stdout.String(), ctx.Err()
	case r := <-done:
		return s.stdout.String(), r.err
	}
}

func (s *Sandbox) Close() error {
	s.loop.Stop()
	return nil
}
