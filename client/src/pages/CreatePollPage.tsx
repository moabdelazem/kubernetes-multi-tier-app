import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreatePoll } from "@/hooks/useCreatePoll";
import { toast } from "sonner";

export default function CreatePollForm() {
  const navigate = useNavigate();
  const [question, setQuestion] = useState("");
  const [options, setOptions] = useState<string[]>(["", ""]);
  const createPollMutation = useCreatePoll();

  const handleAddOption = () => {
    if (options.length < 10) {
      setOptions([...options, ""]);
    }
  };

  const handleRemoveOption = (index: number) => {
    if (options.length > 2) {
      setOptions(options.filter((_, i) => i !== index));
    }
  };

  const handleOptionChange = (index: number, value: string) => {
    const newOptions = [...options];
    newOptions[index] = value;
    setOptions(newOptions);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const validOptions = options.filter((opt) => opt.trim() !== "");

    if (question.trim().length < 5) {
      toast.error("Question must be at least 5 characters long");
      return;
    }

    if (validOptions.length < 2) {
      toast.error("Please provide at least 2 options");
      return;
    }

    try {
      await createPollMutation.mutateAsync({
        question: question.trim(),
        options: validOptions,
      });

      // Show success toast
      toast.success("Poll created successfully!", {
        description: "Your poll is now live and ready for votes.",
      });

      // Reset form
      setQuestion("");
      setOptions(["", ""]);

      // Navigate back to polls list after a short delay
      setTimeout(() => {
        navigate("/");
      }, 1000);
    } catch (error) {
      toast.error("Failed to create poll", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
  };

  return (
    <div className="min-h-screen bg-background p-8">
      <div className="max-w-2xl mx-auto space-y-8">
        <div className="text-center">
          <h1 className="text-4xl font-bold">Create New Poll</h1>
          <p className="text-muted-foreground mt-2">
            Ask a question and provide multiple options
          </p>
        </div>

        <Card className="p-6">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="question">Question</Label>
              <Input
                id="question"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                placeholder="What's your question?"
                className="w-full"
                required
              />
              <p className="text-sm text-muted-foreground">
                Must be between 5 and 500 characters
              </p>
            </div>

            <div className="space-y-4">
              <Label>Options (2-10 required)</Label>
              {options.map((option, index) => (
                <div key={index} className="flex gap-2">
                  <Input
                    value={option}
                    onChange={(e) => handleOptionChange(index, e.target.value)}
                    placeholder={`Option ${index + 1}`}
                    className="flex-1"
                    required
                  />
                  {options.length > 2 && (
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => handleRemoveOption(index)}
                    >
                      Remove
                    </Button>
                  )}
                </div>
              ))}

              {options.length < 10 && (
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleAddOption}
                  className="w-full"
                >
                  + Add Option
                </Button>
              )}
            </div>

            <div className="flex gap-4">
              <Button
                type="submit"
                className="flex-1"
                disabled={createPollMutation.isPending}
              >
                {createPollMutation.isPending ? "Creating..." : "Create Poll"}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate("/")}
              >
                Cancel
              </Button>
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
