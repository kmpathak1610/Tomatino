"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function Dashboard() {
  const [rest, setRest] = useState<any>(null);
  const [stats, setStats] = useState<any>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api("/partner/restaurant").then(setRest).catch((e) => setError("Please login as a restaurant partner."));
    api("/partner/analytics").then(setStats).catch(() => {});
  }, []);

  return (
    <main className="max-w-6xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">{rest ? rest.name : "Restaurant Dashboard"}</h1>
      {error && <p className="text-red-600 mb-4">{error} <a href="/login" className="underline">Login</a></p>}
      {rest && <p className="text-sm text-gray-500 mb-4">Status: {rest.status} · ⭐ {rest.avg_rating} ({rest.total_ratings})</p>}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white rounded shadow p-4"><div className="text-sm text-gray-500">Total Orders</div><div className="text-2xl font-bold">{stats?.orders ?? "-"}</div></div>
        <div className="bg-white rounded shadow p-4"><div className="text-sm text-gray-500">Revenue</div><div className="text-2xl font-bold">₹{stats?.revenue ?? "-"}</div></div>
        <div className="bg-white rounded shadow p-4"><div className="text-sm text-gray-500">Pending Orders</div><div className="text-2xl font-bold text-orange-600">{stats?.pending ?? "-"}</div></div>
      </div>
    </main>
  );
}