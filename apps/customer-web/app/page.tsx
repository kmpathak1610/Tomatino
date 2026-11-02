"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import Link from "next/link";

export default function Home() {
  const [restaurants, setRestaurants] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api("/restaurants")
      .then(setRestaurants)
      .catch(() => setRestaurants([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <main className="max-w-6xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-2">Discover restaurants near you</h1>
      <p className="text-gray-600 mb-6">Order food from the best restaurants in your city.</p>
      {loading && <p>Loading...</p>}
      {!loading && restaurants.length === 0 && <p>No restaurants yet. Seed data via backend.</p>}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
        {restaurants.map((r) => (
          <Link key={r.id} href={`/restaurants/${r.id}`} className="bg-white rounded-lg shadow p-4 hover:shadow-md transition">
            <h2 className="font-semibold text-lg">{r.name}</h2>
            <p className="text-sm text-gray-500">{(r.cuisine_types || []).join(", ")}</p>
            <div className="mt-2 text-sm text-gray-600">⭐ {r.avg_rating} ({r.total_ratings})</div>
            <div className="text-xs text-gray-400 mt-1">{r.status}</div>
          </Link>
        ))}
      </div>
    </main>
  );
}