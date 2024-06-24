"use client";

import { Code, ConnectError } from "@connectrpc/connect";
import { useMutation, useSuspenseQuery } from "@connectrpc/connect-query";
import { Button, Radio, RadioGroup, Spacer } from "@nextui-org/react";
import useTokenHeader from "lib/clerk/token/hook";
import { createRating } from "lib/pb/lmsysmd/rating/v1/rating-RatingService_connectquery";
import {
  type CreateRatingRequest,
  type CreateRatingResponse,
  RatingState_State,
} from "lib/pb/lmsysmd/rating/v1/rating_pb";
import { getSample } from "lib/pb/lmsysmd/sample/v1/sample-SampleService_connectquery";
import {
  type GetSampleRequest,
  type GetSampleResponse,
  type Sample_Choice,
  Sample_Choice_ContentCommon,
} from "lib/pb/lmsysmd/sample/v1/sample_pb";
import { useRouter } from "next/navigation";
import { useQueryState } from "nuqs";
import { type FormEvent, useCallback } from "react";
import { toast } from "sonner";

export default function Rating() {
  const [id, _setId] = useQueryState("id");
  if (!id) throw new Promise((r) => setTimeout(r, 100));
  const sampleId = Number.parseInt(id);
  const tk = useTokenHeader();
  const {
    data: { sample },
  } = useSuspenseQuery<GetSampleRequest, GetSampleResponse>(
    getSample,
    { sampleId },
    { callOptions: { headers: tk } },
  ) as { data: GetSampleResponse };
  const {
    error,
    isError,
    mutateAsync: doCreateRating,
  } = useMutation<CreateRatingRequest, CreateRatingResponse>(createRating, {
    callOptions: { headers: tk },
  });
  const router = useRouter();
  const onSubmit = useCallback(
    async (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      const data = new FormData(e.currentTarget);
      const choice = data.get(sampleId.toString())?.toString();
      if (!choice) {
        toast.error("No choice selected.");
        return;
      }
      if (choice === "skip") {
        router.push("/rating");
        return;
      }
      const choiceId = Number.parseInt(choice);
      const createRatingResponse = doCreateRating({
        rating: { sampleId, choiceId },
        state: { state: RatingState_State.SUBMITTED },
      });
      toast.promise(createRatingResponse, {
        loading: "Submitting Rating...",
        success: ({ ratingId }: CreateRatingResponse) =>
          `Created Rating #${ratingId}.`,
        error: (e: ConnectError) => `Failed to create rating: ${e.message}.`,
      });
      try {
        const { ratingId } = await createRatingResponse;
        const href = `/rating/id/confirm?sid=${sampleId}&cid=${choice}&rid=${ratingId}&ts=${new Date().getTime()}`;
        router.push(href);
      } catch (err) {
        const e = ConnectError.from(err);
        if (e.code === Code.Unauthenticated)
          router.push(`/rating?ts=${new Date().getTime()}`);
        else toast.error(`Something went wrong: ${e.message}.`);
      }
    },
    [doCreateRating, router, sampleId],
  );
  const { content, choices, truth } = sample as {
    content: string;
    choices: Sample_Choice[];
    truth: string;
  };
  return (
    <form className="md:mx-auto md:max-w-md" onSubmit={onSubmit}>
      <p className="font-semibold text-large text-primary">
        Ground Truth: {truth}
      </p>
      <Spacer y={4} />
      <RadioGroup
        classNames={{ label: "text-foreground" }}
        errorMessage={error?.message}
        isInvalid={isError}
        isRequired
        label={content}
        name={sampleId.toString()}
        validationBehavior="native"
      >
        {choices.map(({ choiceId, content }: Sample_Choice, index) => (
          <Radio key={choiceId} value={choiceId.toString()}>
            {content?.value !== Sample_Choice_ContentCommon.NONE_OF_THE_ABOVE &&
              `${index + 1}. `}
            {content?.value !== Sample_Choice_ContentCommon.NONE_OF_THE_ABOVE &&
              content?.value}
            {content?.value === Sample_Choice_ContentCommon.NONE_OF_THE_ABOVE &&
              "None of the above"}
          </Radio>
        ))}
        <Radio value="skip">Skip</Radio>
      </RadioGroup>
      <Spacer y={4} />
      <Button color="primary" fullWidth type="submit">
        Submit&nbsp;Rating
      </Button>
    </form>
  );
}
