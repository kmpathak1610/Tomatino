"use client";
import { useState } from "react";
import { api } from "@/lib/api";
import { useRouter } from "next/navigation";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("owner@spice.com");
  const [password, setPassword] = useState("password123");
  const [error, setError] = useState("");

  const login = async () => {
    try {
      const res: any = await api("/auth/staff/login", { method: "POST", body: JSON.stringify({ email, password }) });
      localStorage.setItem("token", res.access_token);
      router.push("/");
    } catch {
      setError("Invalid credentials");
    }
  };

  return (
    <main className="max-w-md mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Restaurant Login</h1>
      {error && <p className="text-red-600 mb-2">{error}</p>}
      <div className="space-y-3">
        <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" className="w-full border rounded p-2" />
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Password" className="w-full border rounded p-2" />
        <button onClick={login} className="w-full bg-orange-600 text-white py-2 rounded">Login</button>
        <p className="text-xs text-gray-500">Demo: owner@spice.com / password123</p>
      </div>
    </main>
  );
}