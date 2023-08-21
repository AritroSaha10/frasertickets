import auth from '@/lib/firebase/auth'
import { User as firebaseUser, onAuthStateChanged } from 'firebase/auth'
import { ReactNode, createContext, useContext, useEffect, useState } from 'react'

type User = firebaseUser | null
type ContextState = { user: User, loaded: boolean }
type Props = { children?: ReactNode }

const FirebaseAuthContext = createContext<ContextState | undefined>(undefined)

const FirebaseAuthProvider: React.FC<Props> = ({ children }: Props) => {
    const [user, setUser] = useState<User>(null)
    const [loaded, setLoaded] = useState(false)
    const value = { user, loaded }

    useEffect(() => {
        const unsubscribe = onAuthStateChanged(auth, newUser => {
            setUser(newUser)
            setLoaded(true)
        })
        return unsubscribe
    }, [])

    return (
        <FirebaseAuthContext.Provider value={value}>
            {children}
        </FirebaseAuthContext.Provider>
    )
}

function useFirebaseAuth() {
    const context = useContext(FirebaseAuthContext)
    if (context === undefined) {
        throw new Error(
            "useFirebaseAuth must be used within a FirebaseAuthProvider"
        )
    }
    return { user: context.user, loaded: context.loaded }
}

export { FirebaseAuthProvider, useFirebaseAuth }
