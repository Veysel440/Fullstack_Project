import { AuthProvider, useAuth } from "./auth";
import Login from "./pages/Login";
import Items from "./pages/Items";

function Gate() {
    const { user } = useAuth();
    return user ? <Items /> : <Login />;
}

export default function App() {
    return (
        <AuthProvider>
            <Gate />
        </AuthProvider>
    );
}