import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchPolls, fetchPoll, castVote } from "@/services/api";

export function usePolls(limit = 20, offset = 0, activeOnly = true) {
  return useQuery({
    queryKey: ["polls", { limit, offset, activeOnly }],
    queryFn: () => fetchPolls(limit, offset, activeOnly),
  });
}

export function usePoll(pollId: string | undefined) {
  return useQuery({
    queryKey: ["poll", pollId],
    queryFn: () => fetchPoll(pollId!),
    enabled: !!pollId,
  });
}

export function useCastVote() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ pollId, optionId }: { pollId: string; optionId: string }) =>
      castVote(pollId, optionId),
    onSuccess: (_, variables) => {
      // Invalidate and refetch poll data
      queryClient.invalidateQueries({ queryKey: ["poll", variables.pollId] });
      queryClient.invalidateQueries({ queryKey: ["polls"] });
    },
  });
}
