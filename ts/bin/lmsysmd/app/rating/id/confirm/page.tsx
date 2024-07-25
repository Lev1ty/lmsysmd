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
import { type FormEvent, useCallback, useEffect } from "react";
import { toast } from "sonner";
import { useCountdown } from "usehooks-ts";

export default function Confirm() {
  const [sid, _setSid] = useQueryState("sid");
  if (!sid) throw new Promise((r) => setTimeout(r, 100));
  const sampleId = Number.parseInt(sid);
  const [cid, _setCid] = useQueryState("cid");
  if (!cid) throw new Promise((r) => setTimeout(r, 100));
  const [rid, _setRid] = useQueryState("rid");
  if (!rid) throw new Promise((r) => setTimeout(r, 100));
  const tk = useTokenHeader();
  const {
    data: { sample },
  } = useSuspenseQuery<GetSampleRequest, GetSampleResponse>(
    getSample,
    { sampleId },
    { callOptions: { headers: tk } },
  ) as { data: GetSampleResponse };
  const { content, choices, truth } = sample as {
    content: string;
    choices: Sample_Choice[];
    truth: string;
  };
  const {
    error,
    isError,
    mutateAsync: doCreateRating,
  } = useMutation<CreateRatingRequest, CreateRatingResponse>(createRating, {
    callOptions: { headers: tk },
  });
  const [count, { startCountdown }] = useCountdown({
    countStart: 3,
    intervalMs: 1000,
  });
  useEffect(() => {
    startCountdown();
  }, [startCountdown]);
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
      const choiceId = Number.parseInt(choice);
      const ratingId = Number.parseInt(rid);
      const createRatingResponse = doCreateRating({
        rating: { sampleId, choiceId, ratingId },
        state: { state: RatingState_State.CONFIRMED },
      });
      toast.promise(createRatingResponse, {
        loading: "Confirming Rating...",
        success: ({ ratingId }: CreateRatingResponse) =>
          `Confirmed Rating #${ratingId}.`,
        error: (e: ConnectError) => `Failed to confirm rating: ${e.message}.`,
      });
      try {
        await createRatingResponse;
      } catch (err) {
        const e = ConnectError.from(err);
        if (e.code === Code.Unauthenticated)
          router.push(`/rating?ts=${new Date().getTime()}`);
        else toast.error(`Something went wrong: ${e.message}.`);
      }
      router.push(`/rating?ts=${new Date().getTime()}`);
    },
    [doCreateRating, rid, router, sampleId],
  );
  const onPress = useCallback(() => {
    router.back();
  }, [router]);
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
        isReadOnly
        label={content}
        name={sampleId.toString()}
        value={cid}
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
      <div className="flex flex-row gap-2">
        <Button color="primary" fullWidth isLoading={count > 0} type="submit">
          Confirm&nbsp;{!!count && `(${count})`}
        </Button>
        <Button fullWidth onPress={onPress}>
          Go&nbsp;Back
        </Button>
      </div>
    </form>
  );
}
