"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";

const steps = ["PENDING_RESTAURANT_ACCEPTANCE", "ACCEPTED", "PREPARING", "READY_FOR_PICKUP", "PICKED_UP", "OUT_FOR_DELIVERY", "DELIVERED"];

export default function OrderDetailPage() {
  const { id } = useParams();
  const [order, setOrder] = useState<any>(null);

  useEffect(() => {
    if (!id) return;
    const load = () => api(`/orders/${id}`).then(setOrder).catch(() => {});
    load();
    const t = setInterval(load, 5000);
    return () => clearInterval(t);
  }, [id]);

  if (!order) return <main className="p-6">Loading...</main>;

  const idx = steps.indexOf(order.status);

  return (
    <main className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-2">Order {order.order_number}</h1>
      <div className="bg-white rounded shadow p-4 mb-4">
        <div className="font-semibold mb-2">Status: {order.status}</div>
        <div className="text-sm text-gray-600 mb-3">Payment: {order.payment_status}</div>
        <div className="space-y-2">
          {steps.map((s, i) => (
            <div key={s} className={`flex items-center gap-2 text-sm ${i <= idx ? "text-green-600" : "text-gray-400"}`}>
              <span>{i <= idx ? "✓" : "○"}</span>
              <span>{s.replace(/_/g, " ")}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="bg-white rounded shadow p-4">
        <div className="flex justify-between"><span>Subtotal</span><span>₹{order.subtotal}</span></div>
        <div className="flex justify-between font-semibold mt-2"><span>Total</span><span>₹{order.total}</span></div>
      </div>
    </main>
  );
}