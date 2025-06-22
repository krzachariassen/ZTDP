1. We should NEVER allow services to be created without an applicaiton. Services must always belong to an application.

2. We can created a resrouce without a edge to the resource_type in the resource catalog, that should never be allowed.
this is the edges we MUST inforce:
{
		FromKind:     "resource",
		ToKind:       "resource_type",
		AllowedTypes: []string{"instance_of"},
		SpecialRules: validateResourceInstanceToResourceType,
	},

I can also see that we have resource_type created without edge to resource_register, this should also be enforced.
	{
		FromKind:     "resource_register",
		ToKind:       "resource_type",
		AllowedTypes: []string{"owns"},
	},


Deployment agent is using MockAIProvider{} should use real AIProvider interface, not the mock one.