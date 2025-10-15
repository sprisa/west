package policy

import "github.com/sprisa/west/westport/db/ent/privacy"

func DefaultPolicy(policy privacy.Policy) privacy.Policy {
	return privacy.Policy{
		Query: privacy.QueryPolicy{
			policy.Query,
			privacy.AlwaysDenyRule(),
		},
		Mutation: privacy.MutationPolicy{
			policy.Mutation,
			privacy.AlwaysDenyRule(),
		},
	}
}
