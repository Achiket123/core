package hooks

// // HookMappedControl runs on mapped control create and update mutations to restrict certain fields to system admins only
// func HookMappedControl() ent.Hook {
// 	return hook.If(func(next ent.Mutator) ent.Mutator {
// 		return hook.MappedControlFunc(func(ctx context.Context, m *generated.MappedControlMutation) (generated.Value, error) {
// 			if auth.IsSystemAdminFromContext(ctx) {
// 				return next.Mutate(ctx, m)
// 			}

// 			// only system admins can create suggested mappings
// 			mc, ok := m.Source()
// 			if ok && mc == enums.MappingSourceSuggested {
// 				return nil, fmt.Errorf("%w: only system admins can create suggested mappings", ErrInvalidInput)
// 			}

// 			internalID, ok := m.InternalID()
// 			if ok && internalID != "" {
// 				return nil, fmt.Errorf("%w: only system admins can set internal IDs", ErrInvalidInput)
// 			}

// 			internalNotes, ok := m.InternalNotes()
// 			if ok && internalNotes != "" {
// 				return nil, fmt.Errorf("%w: only system admins can set internal notes", ErrInvalidInput)
// 			}

// 			return next.Mutate(ctx, m)
// 		})
// 	},
// 		hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
// 	)
// }
