export const normalizeRole = (role?: string | null): string => {
  return (role || '').toString().trim().toLowerCase();
};

// Public alias for readability in UI code
export const toRoleKey = (role?: string | null): string => normalizeRole(role);

export const normalizeRoles = (roles?: (string | null | undefined)[]): string[] => {
  return (roles || []).map(r => normalizeRole(r)).filter(Boolean);
};

export const isRoleAllowed = (allowedRoles: (string | null | undefined)[] = [], role?: string | null): boolean => {
  const roleNorm = normalizeRole(role);
  const allowed = new Set(normalizeRoles(allowedRoles));
  return roleNorm !== '' && allowed.has(roleNorm);
};

export const humanizeRole = (role?: string | null): string => {
  const r = normalizeRole(role);
  if (!r) return 'Unknown';
  return r.charAt(0).toUpperCase() + r.slice(1);
};

