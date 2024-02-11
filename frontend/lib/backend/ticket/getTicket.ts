import sendBackendRequest from "@/lib/backend/sendBackendRequest";
import { convertToTicketWithEventData } from "@/lib/backend/ticket/ticketWithUserAndEventData";

export default async function getTicket(id: string) {
    const res = await sendBackendRequest(`/tickets/${id}`, "get");
    const ticket = convertToTicketWithEventData(res.data as { [key: string]: any });
    return ticket;
}
