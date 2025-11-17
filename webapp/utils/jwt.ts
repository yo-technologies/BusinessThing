// Utility to decode JWT and extract organizations
export interface Organization {
  id: string;
  role: string;
}

export interface JWTPayload {
  sub: string; // user_id
  orgs: Organization[];
  iat: number;
  exp: number;
}

export function decodeJWT(token: string): JWTPayload | null {
  try {
    const parts = token.split(".");

    if (parts.length !== 3) {
      return null;
    }

    const payload = parts[1];
    const decoded = JSON.parse(atob(payload));

    return {
      sub: decoded.sub || "",
      orgs: decoded.orgs || [],
      iat: decoded.iat || 0,
      exp: decoded.exp || 0,
    };
  } catch (error) {
    console.error("Failed to decode JWT:", error);

    return null;
  }
}

export function getOrganizationsFromToken(
  token: string | null,
): Organization[] {
  if (!token) return [];

  const payload = decodeJWT(token);

  return payload?.orgs || [];
}
