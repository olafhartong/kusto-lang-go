package kustolang

// Ported from:
//   src/Kusto.Language/Editor/CodeServiceFactory.cs
//   src/Kusto.Language/Editor/OffsetCodeService.cs
//   src/Kusto.Language/Editor/UnknownCodeService.cs
//   src/Kusto.Language/Editor/Kusto/KustoCodeService.cs
//   src/Kusto.Language/Editor/Kusto/KustoCodeServiceFactory.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// CommonCodeService exposes the reusable behavior of the upstream abstract
// editor service through Go composition.
type CommonCodeService struct {
	*CodeService
}

func NewCommonCodeService(text string) *CommonCodeService {
	return &CommonCodeService{CodeService: wrapCodeService(internal.NewCommonCodeService(text))}
}

func wrapCommonCodeService(service internal.CodeService) *CommonCodeService {
	return &CommonCodeService{CodeService: wrapCodeService(service)}
}

// KustoCodeService is the concrete Kusto editor service.
type KustoCodeService struct {
	*CommonCodeService
}

func NewKustoCodeService(text string, globals ...*GlobalState) *KustoCodeService {
	state := DefaultGlobalState
	if len(globals) > 0 && globals[0] != nil {
		state = globals[0]
	}
	return &KustoCodeService{CommonCodeService: wrapCommonCodeService(internal.NewKustoCodeService(text, state))}
}

func NewKustoCodeServiceForCode(code *KustoCode) *KustoCodeService {
	var inner *internal.KustoCode
	if code != nil {
		inner = code.code
	}
	return &KustoCodeService{CommonCodeService: wrapCommonCodeService(internal.NewKustoCodeServiceForCode(inner))}
}

// UnknownCodeService provides the source-compatible fallback editor behavior.
type UnknownCodeService struct {
	*CommonCodeService
}

func NewUnknownCodeService(text string) *UnknownCodeService {
	return &UnknownCodeService{CommonCodeService: wrapCommonCodeService(internal.NewUnknownCodeService(text))}
}

// OffsetCodeService adjusts all public editor positions by a fixed UTF-16 offset.
type OffsetCodeService struct {
	*CodeService
	offsetService *internal.OffsetCodeService
}

func NewOffsetCodeService(service *CodeService, offset int) *OffsetCodeService {
	var inner internal.CodeService
	if service != nil {
		inner = service.service
	}
	offsetService := internal.NewOffsetCodeService(inner, offset)
	return &OffsetCodeService{
		CodeService:   wrapCodeService(offsetService),
		offsetService: offsetService,
	}
}

func (s *OffsetCodeService) WithOffset(offset int) *OffsetCodeService {
	if s == nil || s.offsetService == nil {
		return NewOffsetCodeService(nil, offset)
	}
	updated := s.offsetService.WithOffset(offset)
	return &OffsetCodeService{CodeService: wrapCodeService(updated), offsetService: updated}
}

// CodeServiceFactory maps the upstream abstract factory to a composable Go
// callback. Factories with the same key replace one another when combined.
type CodeServiceFactory struct {
	factory internal.CodeServiceFactory
	key     string
}

func NewCodeServiceFactory(key string, try func(string) (*CodeService, bool)) *CodeServiceFactory {
	adapter := internal.NewCodeServiceFactoryFunc(key, func(text string) (internal.CodeService, bool) {
		if try == nil {
			return nil, false
		}
		service, ok := try(text)
		if !ok || service == nil {
			return nil, false
		}
		return service.service, true
	})
	return &CodeServiceFactory{factory: adapter, key: key}
}

func wrapCodeServiceFactory(factory internal.CodeServiceFactory) *CodeServiceFactory {
	if factory == nil {
		return nil
	}
	return &CodeServiceFactory{factory: factory, key: internal.CodeServiceFactoryKey(factory)}
}

func (f *CodeServiceFactory) Key() string {
	if f == nil {
		return ""
	}
	return f.key
}

func (f *CodeServiceFactory) TryGetCodeService(text string) (*CodeService, bool) {
	if f == nil || f.factory == nil {
		return nil, false
	}
	service, ok := f.factory.TryGetCodeService(text)
	if !ok || service == nil {
		return nil, false
	}
	return wrapCodeService(service), true
}

func (f *CodeServiceFactory) GetFactory(match func(*CodeServiceFactory) bool) *CodeServiceFactory {
	if f == nil || f.factory == nil {
		return nil
	}
	found := f.factory.GetFactory(func(candidate internal.CodeServiceFactory) bool {
		return match == nil || match(wrapCodeServiceFactory(candidate))
	})
	return wrapCodeServiceFactory(found)
}

// GetKustoCodeServiceFactory returns the concrete Kusto factory contained in
// this factory, including when this factory is an aggregate.
func (f *CodeServiceFactory) GetKustoCodeServiceFactory() *KustoCodeServiceFactory {
	if f == nil || f.factory == nil {
		return nil
	}
	found, _ := f.factory.GetFactory(func(candidate internal.CodeServiceFactory) bool {
		_, ok := candidate.(*internal.KustoCodeServiceFactory)
		return ok
	}).(*internal.KustoCodeServiceFactory)
	return wrapKustoCodeServiceFactory(found)
}

func (f *CodeServiceFactory) WithFactory(factory *CodeServiceFactory) *CodeServiceFactory {
	if f == nil || f.factory == nil {
		return factory
	}
	if factory == nil || factory.factory == nil {
		return f
	}
	return wrapCodeServiceFactory(f.factory.WithFactory(factory.factory))
}

func NewUnknownCodeServiceFactory() *CodeServiceFactory {
	return wrapCodeServiceFactory(internal.NewUnknownCodeServiceFactory())
}

// KustoCodeServiceFactory creates Kusto services and rejects SQL-looking text,
// matching the upstream factory heuristic.
type KustoCodeServiceFactory struct {
	*CodeServiceFactory
	factory *internal.KustoCodeServiceFactory
}

func NewKustoCodeServiceFactory(globals *GlobalState) *KustoCodeServiceFactory {
	return wrapKustoCodeServiceFactory(internal.NewKustoCodeServiceFactory(globals))
}

func wrapKustoCodeServiceFactory(factory *internal.KustoCodeServiceFactory) *KustoCodeServiceFactory {
	if factory == nil {
		return nil
	}
	return &KustoCodeServiceFactory{
		CodeServiceFactory: wrapCodeServiceFactory(factory),
		factory:            factory,
	}
}

func (f *KustoCodeServiceFactory) Globals() *GlobalState {
	if f == nil || f.factory == nil {
		return nil
	}
	return f.factory.Globals
}

func (f *KustoCodeServiceFactory) WithGlobals(globals *GlobalState) *KustoCodeServiceFactory {
	if f == nil || f.factory == nil {
		return NewKustoCodeServiceFactory(globals)
	}
	updated := f.factory.WithGlobals(globals)
	if updated == f.factory {
		return f
	}
	return wrapKustoCodeServiceFactory(updated)
}
