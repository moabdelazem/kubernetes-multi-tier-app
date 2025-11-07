import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createPoll } from "@/services/api";
import type { CreatePollRequest } from "@/types/poll";

export function useCreatePoll() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (poll: CreatePollRequest) => createPoll(poll),
    onSuccess: () => {
      // Invalidate polls list to refetch
      queryClient.invalidateQueries({ queryKey: ["polls"] });
    },
  });
}
