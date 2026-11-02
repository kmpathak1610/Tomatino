"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function MenuPage() {
  const [items, setItems] = useState<any[]>([]);
  const [error, setError] = useState("");

  const load = () => api("/partner/menu").then(setItems).catch(() => setError("Login required"));
  useEffect(load, []);

  const toggle = async (item: any) => {
    await api(`/partner/menu/${item.id}/availability`, { method: "PATCH", body: JSON.stringify({ available: !item.is_available }) });
    load();
  };

  return (
    <main className="max-w-5xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Menu Management</h1>
      {error && <p className="text-red-600 mb-3">{error}</p>}
      <div className="bg-white rounded shadow divide-y">
        {items.map((i) => (
          <div key={i.id} className="p-4 flex items-center justify-between">
            <div>
              <div className="font-medium">{i.is_veg ? "🟢" : "🔴"} {i.name}</div>
              <div className="text-sm text-gray-500">₹{i.base_price}</div>
            </div>
            <button onClick={() => toggle(i)} className={`px-3 py-1 rounded text-sm text-white ${i.is_available ? "bg-green-600" : "bg-gray-400"}`}>
              {i.is_available ? "Available" : "Unavailable"}
            </button>
          </div>
        ))}
        {items.length === 0 && <div className="p-4 text-gray-500">No menu items.</div>}
      </div>
    </main>
  );
}