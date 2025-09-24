import { useState } from "react";
import { useAuth } from "../auth";

export default function Login() {
    const { login } = useAuth();
    const [email, setEmail] = useState("admin@example.com");
    const [password, setPassword] = useState("admin123");
    const [err, setErr] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    async function submit(e: React.FormEvent) {
        e.preventDefault();
        setLoading(true);
        setErr(null);
        const ok = await login(email, password);
        if (!ok) setErr("Invalid credentials");
        setLoading(false);
    }

    return (
        <div style={{ maxWidth: 360, margin: "80px auto", fontFamily: "system-ui" }}>
            <h2>Sign in</h2>
            <form onSubmit={submit} style={{ display: "grid", gap: 8 }}>
                <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email" />
                <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="password" />
                <button disabled={loading} type="submit">Login</button>
            </form>
            {err && <p style={{ color: "crimson" }}>{err}</p>}
        </div>
    );
}
