import { useAuth } from "@clerk/clerk-react";
import { useSuspenseQuery } from "@tanstack/react-query";

export default function useTokenHeader() {
  const { getToken, isLoaded } = useAuth();
  if (!isLoaded) throw new Promise((r) => setTimeout(r, 100));
  const { data } = useSuspenseQuery({
    queryKey: ["clerk"],
    queryFn: async () => {
      const token = await getToken();
      if (!token) throw new Error("null clerk auth token");
      return token;
    },
    staleTime: 0,
  });
  return { authorization: `Bearer ${data}` };
}
