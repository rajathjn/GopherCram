export interface User {
  id: number;
  name: string;
  email: string;
}

// fetchUsers is a stub that pretends to call a remote API. The fixture is
// entirely synthetic and contains no real endpoints or credentials.
export async function fetchUsers(): Promise<User[]> {
  await new Promise((r) => setTimeout(r, 1));
  return [
    { id: 1, name: "Ada Lovelace", email: "ada@example.test" },
    { id: 2, name: "Grace Hopper", email: "grace@example.test" },
    { id: 3, name: "Linus Torvalds", email: "linus@example.test" },
  ];
}

export function findUserByEmail(users: User[], email: string): User | undefined {
  return users.find((u) => u.email.toLowerCase() === email.toLowerCase());
}
