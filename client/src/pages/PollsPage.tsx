import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import Poll from "@/components/poll";
import { usePolls, useCastVote } from "@/hooks/usePolls";

export default function PollsPage() {
  const { data: polls, isLoading, error } = usePolls();
  const castVoteMutation = useCastVote();

  const handleVote = (pollId: string, optionId: string) => {
    castVoteMutation.mutate({ pollId, optionId });
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background p-8">
        <div className="max-w-4xl mx-auto">
          <div className="text-center">Loading polls...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-background p-8">
        <div className="max-w-4xl mx-auto">
          <div className="text-center text-destructive">
            Error loading polls: {error.message}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-4xl font-bold">Quick Polls</h1>
            <p className="text-muted-foreground mt-2">
              Vote on polls and see live results
            </p>
          </div>
          <Link to="/create">
            <Button variant="outline" size="lg">
              + Create Poll
            </Button>
          </Link>
        </div>

        {/* Polls List */}
        {polls && polls.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {polls.map((poll) => (
              <Poll
                key={poll.id}
                question={poll.question}
                options={
                  poll.options?.map((opt) => ({
                    id: opt.id,
                    label: opt.option_text,
                    votes: opt.vote_count,
                  })) || []
                }
                onVote={(optionId) => handleVote(poll.id, optionId)}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-12">
            <p className="text-muted-foreground mb-4">No polls available yet</p>
            <Link to="/create">
              <Button>Create the first poll</Button>
            </Link>
          </div>
        )}

        {castVoteMutation.isPending && (
          <div className="fixed bottom-4 right-4 bg-primary text-primary-foreground px-4 py-2 rounded-lg shadow-lg">
            Submitting vote...
          </div>
        )}
      </div>
    </div>
  );
}
