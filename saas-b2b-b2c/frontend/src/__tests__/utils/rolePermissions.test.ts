import { getAssignableRoles, canSeeGoal, assignableRolesMap } from '@/utils/rolePermissions';

describe('rolePermissions', () => {
  describe('assignableRolesMap', () => {
    it('super_admin can assign franchiser', () => {
      expect(assignableRolesMap['super_admin']).toContain('franchiser');
    });

    it('super_admin can assign dealer', () => {
      expect(assignableRolesMap['super_admin']).toContain('dealer');
    });

    it('franchiser can assign dealer', () => {
      expect(assignableRolesMap['franchiser']).toContain('dealer');
    });

    it('franchiser can assign franchise_manager', () => {
      expect(assignableRolesMap['franchiser']).toContain('franchise_manager');
    });

    it('franchiser_manager can assign dealer', () => {
      expect(assignableRolesMap['franchiser_manager']).toContain('dealer');
    });

    it('franchiser_manager can assign salon_manager', () => {
      expect(assignableRolesMap['franchiser_manager']).toContain('salon_manager');
    });

    it('dealer can only assign salon_manager', () => {
      expect(assignableRolesMap['dealer']).toEqual(['salon_manager']);
    });

    it('salon_manager cannot assign anyone', () => {
      expect(assignableRolesMap['salon_manager']).toEqual([]);
    });
  });

  describe('getAssignableRoles', () => {
    it('returns roles for super_admin', () => {
      const roles = getAssignableRoles('super_admin');
      expect(roles).toContain('franchiser');
      expect(roles).toContain('dealer');
      expect(roles).toContain('franchiser_manager');
      expect(roles).toContain('salon_manager');
    });

    it('returns roles for franchiser', () => {
      const roles = getAssignableRoles('franchiser');
      expect(roles).toContain('dealer');
      expect(roles).toContain('franchise_manager');
      expect(roles).toContain('salon_manager');
    });

    it('returns roles for franchiser_manager', () => {
      const roles = getAssignableRoles('franchiser_manager');
      expect(roles).toContain('dealer');
      expect(roles).toContain('salon_manager');
      expect(roles).not.toContain('franchiser');
    });

    it('returns roles for dealer', () => {
      const roles = getAssignableRoles('dealer');
      expect(roles).toEqual(['salon_manager']);
    });

    it('returns empty array for salon_manager', () => {
      const roles = getAssignableRoles('salon_manager');
      expect(roles).toEqual([]);
    });

    it('returns empty array for null role', () => {
      const roles = getAssignableRoles(null);
      expect(roles).toEqual([]);
    });

    it('returns empty array for undefined role', () => {
      const roles = getAssignableRoles(undefined);
      expect(roles).toEqual([]);
    });

    it('returns empty array for unknown role', () => {
      const roles = getAssignableRoles('unknown_role');
      expect(roles).toEqual([]);
    });

    it('returns empty array for empty string', () => {
      const roles = getAssignableRoles('');
      expect(roles).toEqual([]);
    });
  });

  describe('canSeeGoal', () => {
    it('allows seeing own goal', () => {
      const canSee = canSeeGoal('franchiser', 'dealer', 'user-1', 'user-1');
      expect(canSee).toBe(true);
    });

    it('franchiser can see dealer goal', () => {
      const canSee = canSeeGoal('franchiser', 'dealer', 'user-1', 'user-2');
      expect(canSee).toBe(true);
    });

    it('franchiser cannot see franchiser goal of another user', () => {
      const canSee = canSeeGoal('franchiser', 'franchiser', 'user-1', 'user-2');
      expect(canSee).toBe(false);
    });

    it('dealer can see own salon_manager goal', () => {
      const canSee = canSeeGoal('dealer', 'salon_manager', 'user-1', 'user-2');
      expect(canSee).toBe(true);
    });

    it('dealer cannot see other dealer goal', () => {
      const canSee = canSeeGoal('dealer', 'dealer', 'user-1', 'user-2');
      expect(canSee).toBe(false);
    });

    it('salon_manager cannot see any goal', () => {
      const canSee = canSeeGoal('salon_manager', 'dealer', 'user-1', 'user-2');
      expect(canSee).toBe(false);
    });

    it('null myRole cannot see any goal', () => {
      const canSee = canSeeGoal(null, 'dealer', 'user-1', 'user-2');
      expect(canSee).toBe(false);
    });

    it('same user ID always sees goal regardless of role', () => {
      const canSee = canSeeGoal('salon_manager', 'dealer', 'user-1', 'user-1');
      expect(canSee).toBe(true);
    });
  });
});

describe('Role Hierarchy', () => {
  it('has correct role hierarchy count', () => {
    const roles = Object.keys(assignableRolesMap);
    expect(roles).toHaveLength(5);
  });

  it('franchiser has most assignable roles', () => {
    const franchiserRoles = getAssignableRoles('franchiser');
    const dealerRoles = getAssignableRoles('dealer');
    expect(franchiserRoles.length).toBeGreaterThan(dealerRoles.length);
  });

  it('each role has decreasing permissions', () => {
    const superAdminCount = getAssignableRoles('super_admin').length;
    const franchiserCount = getAssignableRoles('franchiser').length;
    const managerCount = getAssignableRoles('franchiser_manager').length;
    const dealerCount = getAssignableRoles('dealer').length;
    const salonCount = getAssignableRoles('salon_manager').length;

    expect(superAdminCount).toBe(4);
    expect(franchiserCount).toBe(3);
    expect(managerCount).toBe(2);
    expect(dealerCount).toBe(1);
    expect(salonCount).toBe(0);
  });
});