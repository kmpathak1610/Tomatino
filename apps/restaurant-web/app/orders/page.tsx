"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

const nextStatus: Record<string, string> = {
  PENDING_RESTAURANT_ACCEPTANCE: "ACCEPTED",
  ACCEPTED: "PREPARING",
  PREPARING: "READY_FOR_PICKUP",
};

export default function OrdersPage() {
  const [orders, setOrders] = useState<any[]>([]);
  const [error, setError] = useState("");

  const load = () => api("/partner/orders").then(setOrders).catch(() => setError("Login required"));
  useEffect(() => { load(); const t = setInterval(load, 5000); return () => clearInterval(t); }, []);

  const update = async (id: string, status: string) => {
    try {
      await api(`/orders/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) });
      load();
    } catch (e: any) { alert("Failed: " + e.message); }
  };

  return (
    <main className="max-w-5xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Incoming Orders</h1>
      {error && <p className="text-red-600 mb-3">{error}</p>}
      {orders.length === 0 && <p className="text-gray-500">No orders yet.</p>}
      <div className="space-y-3">
        {orders.map((o) => (
          <div key={o.id} className="bg-white rounded shadow p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <div className="font-semibold">{o.order_number}</div>
              <div className="text-sm text-gray-500">{o.status} · {o.payment_status}</div>
              <div className="text-sm">₹{o.total}</div>
            </div>
            <div className="flex gap-2">
              {nextStatus[o.status] && (
                <button onClick={() => update(o.id, nextStatus[o.status])} className="bg-green-600 text-white px-3 py-1 rounded text-sm">
                  Mark {nextStatus[o.status].replace(/_/g, " ")}
                </button>
              )}
              {o.status === "PENDING_RESTAURANT_ACCEPTANCE" && (
                <button onClick={() => update(o.id, "REJECTED")} className="bg-red-600 text-white px-3 py-1 rounded text-sm">Reject</button>
              )}
            </div>
          </div>
        ))}
      </div>
    </main>
  );
}