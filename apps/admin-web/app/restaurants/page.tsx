"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function RestaurantsPage() {
  const [rests, setRests] = useState<any[]>([]);
  const load = () => api("/admin/restaurants").then(setRests).catch(() => {});
  useEffect(load, []);
  const approve = async (id: string) => { await api(`/restaurants/${id}/approve`, { method: "POST" }); load(); };

  return (
    <main className="max-w-6xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Restaurants</h1>
      <div className="space-y-2">
        {rests.map((r) => (
          <div key={r.id} className="bg-white rounded shadow p-4 flex justify-between items-center">
            <div><div className="font-medium">{r.name}</div><div className="text-sm text-gray-500">{r.status} · ⭐ {r.avg_rating}</div></div>
            {r.status !== "APPROVED" && <button onClick={() => approve(r.id)} className="bg-green-600 text-white px-3 py-1 rounded text-sm">Approve</button>}
          </div>
        ))}
        {rests.length === 0 && <p className="text-gray-500">No restaurants.</p>}
      </div>
    </main>
  );
}