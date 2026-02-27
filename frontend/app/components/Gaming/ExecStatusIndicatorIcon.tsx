import {
	faCircle,
	faCircleCheck,
	faCircleExclamation,
	faRotate,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import type { components } from "../../api/schema";

type Props = {
	status: components["schemas"]["ExecutionStatus"];
};

export default function ExecStatusIndicatorIcon({ status }: Props) {
	switch (status) {
		case "none":
			return (
				<FontAwesomeIcon icon={faCircle} fixedWidth className="text-gray-400" />
			);
		case "running":
			return (
				<FontAwesomeIcon
					icon={faRotate}
					spin
					fixedWidth
					className="text-gray-700"
				/>
			);
		case "success":
			return (
				<FontAwesomeIcon
					icon={faCircleCheck}
					fixedWidth
					className="text-brand-500"
				/>
			);
		default:
			return (
				<FontAwesomeIcon
					icon={faCircleExclamation}
					fixedWidth
					className="text-red-500"
				/>
			);
	}
}
