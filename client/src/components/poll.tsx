"use client";

import { useState } from "react";
import { Card } from "@/components/ui/card";

interface PollOption {
  id: string;
  label: string;
  votes: number;
}

interface PollProps {
  question: string;
  options: PollOption[];
  onVote?: (optionId: string) => void;
}

export default function Poll({ question, options, onVote }: PollProps) {
  const [voted, setVoted] = useState<string | null>(null);
  const [pollOptions, setPollOptions] = useState(options);

  const totalVotes = pollOptions.reduce((sum, option) => sum + option.votes, 0);

  const handleVote = (optionId: string) => {
    if (voted) return;

    setVoted(optionId);
    setPollOptions(
      pollOptions.map((option) =>
        option.id === optionId ? { ...option, votes: option.votes + 1 } : option
      )
    );
    onVote?.(optionId);
  };

  return (
    <Card className="w-full max-w-md bg-card p-6">
      <div className="space-y-6">
        {/* Poll Question */}
        <div>
          <h3 className="text-lg font-semibold text-card-foreground">
            {question}
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {totalVotes} votes so far
          </p>
        </div>

        {/* Poll Options */}
        <div className="space-y-3">
          {pollOptions.map((option) => {
            const percentage =
              totalVotes > 0
                ? Math.round((option.votes / totalVotes) * 100)
                : 0;
            const isSelected = voted === option.id;
            const canVote = !voted;

            return (
              <button
                key={option.id}
                onClick={() => handleVote(option.id)}
                disabled={!canVote}
                className={`w-full text-left transition-all duration-200 ${
                  canVote ? "cursor-pointer" : "cursor-default"
                }`}
                aria-label={`Vote for ${option.label}`}
              >
                <div className="flex items-end justify-between gap-2 mb-2">
                  <span
                    className={`text-sm font-medium transition-colors ${
                      isSelected ? "text-primary" : "text-foreground"
                    }`}
                  >
                    {option.label}
                  </span>
                  <span className="text-xs text-muted-foreground font-medium">
                    {percentage}%
                  </span>
                </div>

                {/* Progress Bar */}
                <div className="relative h-2 w-full overflow-hidden rounded-full bg-muted">
                  <div
                    className={`h-full transition-all duration-500 ease-out ${
                      isSelected ? "bg-primary" : "bg-accent"
                    }`}
                    style={{ width: `${percentage}%` }}
                    role="progressbar"
                    aria-valuenow={percentage}
                    aria-valuemin={0}
                    aria-valuemax={100}
                  />
                </div>

                {/* Vote Count */}
                <p className="mt-1 text-xs text-muted-foreground">
                  {option.votes} vote{option.votes !== 1 ? "s" : ""}
                </p>
              </button>
            );
          })}
        </div>

        {/* Voted State */}
        {voted && (
          <div className="flex items-center justify-center rounded-lg bg-muted/50 py-2 px-3">
            <p className="text-sm font-medium text-muted-foreground">
              ✓ Vote submitted
            </p>
          </div>
        )}
      </div>
    </Card>
  );
}
