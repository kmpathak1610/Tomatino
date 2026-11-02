"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function UsersPage() {
  const [users, setUsers] = useState<any[]>([]);
  useEffect(() => { api("/admin/users").then(setUsers).catch(() => {}); }, []);
  return (
    <main className="max-w-6xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Customers & Users</h1>
      <div className="bg-white rounded shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-100 text-left"><tr><th className="p-3">Phone</th><th className="p-3">Email</th><th className="p-3">Status</th></tr></thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id} className="border-t"><td className="p-3">{u.phone || "-"}</td><td className="p-3">{u.email || "-"}</td><td className="p-3">{u.status}</td></tr>
            ))}
            {users.length === 0 && <tr><td className="p-3 text-gray-500" colSpan={3}>No users.</td></tr>}
          </tbody>
        </table>
      </div>
    </main>
  );
}