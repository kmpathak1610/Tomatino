"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function EarningsPage() {
  const [mine, setMine] = useState<any[]>([]);
  useEffect(() => { api("/delivery/assignments").then(setMine).catch(() => {}); }, []);
  const done = mine.filter((a) => a.status === "DELIVERED").length;
  return (
    <main className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Earnings</h1>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white rounded shadow p-4"><div className="text-sm text-gray-500">Completed Deliveries</div><div className="text-2xl font-bold">{done}</div></div>
        <div className="bg-white rounded shadow p-4"><div className="text-sm text-gray-500">Estimated Earnings</div><div className="text-2xl font-bold">₹{done * 40}</div></div>
        <div className="bg-white rounded shadow p-4"><div className="text-sm text-gray-500">Incentives</div><div className="text-2xl font-bold">₹{done >= 5 ? 100 : 0}</div></div>
      </div>
    </main>
  );
}