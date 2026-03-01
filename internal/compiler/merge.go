package compiler

// Merge combines multiple pack models in order. Later packs append policies, prompts, rules;
// targets are merged by id (later overwrites). Pack name/version are taken from the first model.
func Merge(models []*Model) *Model {
	if len(models) == 0 {
		return &Model{Targets: make(map[string]TargetConfig)}
	}
	out := &Model{
		Pack:    models[0].Pack,
		Targets: make(map[string]TargetConfig),
	}
	for _, m := range models {
		out.Policies = append(out.Policies, m.Policies...)
		out.Prompts = append(out.Prompts, m.Prompts...)
		out.Rules = append(out.Rules, m.Rules...)
		for id, tc := range m.Targets {
			out.Targets[id] = tc
		}
	}
	return out
}
