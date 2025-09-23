import { useEffect, useState } from "react";
type Item = { id: number; name: string; price: number; created_at: string };
const API = (import.meta.env.VITE_API_URL as string) ?? "http://localhost:8080";

export default function App() {
    const [items, setItems] = useState<Item[]>([]);
    const [name, setName] = useState(""); const [price, setPrice] = useState(0);

    async function load() {
        const res = await fetch(`${API}/items/`);
        setItems(await res.json());
    }
    async function add(e: React.FormEvent) {
        e.preventDefault();
        await fetch(`${API}/items/`, {
            method: "POST", headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name, price: Number(price) })
        });
        setName(""); setPrice(0); await load();
    }
    useEffect(() => { void load(); }, []);

    return (
        <div style={{maxWidth:820,margin:"40px auto",fontFamily:"system-ui"}}>
            <h1>PostgreSQL Items</h1>
            <form onSubmit={add} style={{display:"flex",gap:8}}>
                <input value={name} onChange={e=>setName(e.target.value)} placeholder="Ad" required/>
                <input type="number" step="0.01" value={price} onChange={e=>setPrice(Number(e.target.value))} placeholder="Fiyat"/>
                <button type="submit">Ekle</button>
            </form><hr/>
            <ul>{items.map(i=>(
                <li key={i.id}>#{i.id} {i.name} — {i.price.toFixed(2)} — {new Date(i.created_at).toLocaleString()}</li>
            ))}</ul>
        </div>
    );
}
