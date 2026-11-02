"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function RiderDashboard() {
  const [available, setAvailable] = useState<any[]>([]);
  const [mine, setMine] = useState<any[]>([]);
  const [error, setError] = useState("");

  const load = () => {
    api("/delivery/available").then(setAvailable).catch(() => setError("Login required"));
    api("/delivery/assignments").then(setMine).catch(() => {});
  };
  useEffect(() => { load(); const t = setInterval(load, 5000); return () => clearInterval(t); }, []);

  const claim = async (orderId: string) => {
    await api(`/delivery/claim/${orderId}`, { method: "POST" });
    load();
  };
  const act = async (id: string, action: string) => {
    await api(`/delivery/assignments/${id}/${action}`, { method: "POST" });
    load();
  };
  const complete = async (id: string) => {
    const otp = prompt("Enter customer delivery OTP (dev: 1234)");
    if (!otp) return;
    await api(`/delivery/assignments/${id}/complete`, { method: "POST", body: JSON.stringify({ otp }) });
    load();
  };

  return (
    <main className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Delivery Portal</h1>
      {error && <p className="text-red-600 mb-3">{error} <a href="/login" className="underline">Login</a></p>}

      <h2 className="text-lg font-semibold mb-2">Available Requests</h2>
      <div className="space-y-2 mb-6">
        {available.map((o) => (
          <div key={o.order_id} className="bg-white rounded shadow p-4 flex justify-between items-center">
            <div><div className="font-medium">{o.order_number}</div><div className="text-sm text-gray-500">{o.status} · ₹{o.total}</div></div>
            <button onClick={() => claim(o.order_id)} className="bg-orange-600 text-white px-3 py-1 rounded text-sm">Accept</button>
          </div>
        ))}
        {available.length === 0 && <p className="text-gray-500 text-sm">No available requests.</p>}
      </div>

      <h2 className="text-lg font-semibold mb-2">My Deliveries</h2>
      <div className="space-y-2">
        {mine.map((a) => (
          <div key={a.id} className="bg-white rounded shadow p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div><div className="font-medium">{a.order_number}</div><div className="text-sm text-gray-500">{a.status} · {a.order_status}</div></div>
            <div className="flex gap-2">
              {a.status === "ASSIGNED" && <button onClick={() => act(a.id, "accept")} className="bg-blue-600 text-white px-3 py-1 rounded text-sm">Accept</button>}
              {a.status === "ACCEPTED" && <button onClick={() => act(a.id, "pickup")} className="bg-indigo-600 text-white px-3 py-1 rounded text-sm">Picked Up</button>}
              {a.status === "PICKED_UP" && <button onClick={() => act(a.id, "out-for-delivery")} className="bg-purple-600 text-white px-3 py-1 rounded text-sm">Out for Delivery</button>}
              {(a.status === "PICKED_UP" || a.status === "OUT_FOR_DELIVERY") && <button onClick={() => complete(a.id)} className="bg-green-600 text-white px-3 py-1 rounded text-sm">Complete + OTP</button>}
            </div>
          </div>
        ))}
        {mine.length === 0 && <p className="text-gray-500 text-sm">No deliveries assigned.</p>}
      </div>
    </main>
  );
}