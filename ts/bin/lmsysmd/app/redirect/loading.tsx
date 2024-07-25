import { Spinner } from "@nextui-org/react";

export default function RedirectLoading() {
  return (
    <div className="flex flex-col justify-center items-center w-full h-full">
      <Spinner color="primary" size="lg" />
    </div>
  );
}
