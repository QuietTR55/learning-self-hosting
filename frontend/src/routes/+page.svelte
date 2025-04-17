<script lang="ts">
    import { onMount, tick } from "svelte";

    let messages = $state<string[]>([]);
    let ws = $state<WebSocket | null>(null);
    let readyState = $state<number | null>(null);

    let message = $state("");

    let messageContainer: HTMLDivElement | null = null;
    let userScrolledUp = $state(false);
    const scrollThreshold = 50;

    // Optional: Map numbers to strings for better display
    const stateMap: { [key: number]: string } = {
        0: "CONNECTING",
        1: "OPEN",
        2: "CLOSING",
        3: "CLOSED"
    };
    let readyStateString: string = $derived(readyState !== null ? (stateMap[readyState] || "UNKNOWN") : "DISCONNECTED");

    function isNearBottom(element: HTMLDivElement | null): boolean {
        if (!element) return true;
        return element.scrollHeight - element.scrollTop - element.clientHeight < scrollThreshold;
    }

    function handleScroll() {
        if (!messageContainer) return;
        if (isNearBottom(messageContainer)) {
            userScrolledUp = false;
        } else {
            userScrolledUp = true;
        }
    }

    const sendMessage = async () => {
        const formData = new FormData();
        formData.append("message", message);
        try {
            const result = await fetch("http://localhost:8080/chat", {
                method: "POST",
                body: formData
            });
            if (result.ok) {
                message = ""; // Clear input only on success
            } else {
                console.error("Failed to send message:", result.status);
                // Handle error display if needed
            }
        } catch (error) {
            console.error("Error sending message:", error);
        }
    }

    onMount(() => {
        console.log("Mounted");
        const socket = new WebSocket("ws://localhost:8080/ws");
        ws = socket; // Assign to non-reactive variable
        readyState = socket.readyState; // Set initial reactive state

        const fetchMessages = async () => {
            try {
                const result = await fetch("http://localhost:8080/chat");
                if (result.ok) {
                    const data = await result.json();
                    messages = data.map((msg: any) => msg.Content).reverse();
                    // Scroll after initial fetch
                    await tick();
                    if (messageContainer) {
                        messageContainer.scrollTop = messageContainer.scrollHeight;
                    }
                    userScrolledUp = false;
                } else {
                    console.error("Failed to fetch initial messages:", result.status);
                }
            } catch(error) {
                 console.error("Error fetching initial messages:", error);
            }
        }

        fetchMessages();

        socket.onopen = () => {
            console.log("Connected to server");
            readyState = socket.readyState; // Update reactive state
        };

        socket.onmessage = async (event) => {
            console.log("Message from server ", event.data);

            const shouldScroll = isNearBottom(messageContainer);

            messages.push(event.data);
            messages = messages;

            await tick();

            if (shouldScroll && !userScrolledUp && messageContainer) {
                messageContainer.scrollTop = messageContainer.scrollHeight;
            }
        };

        socket.onclose = (event) => {
            console.log("Disconnected from server:", event.code, event.reason);
            readyState = socket.readyState; // Update reactive state
            ws = null; // Clear the reference if needed
        };

        socket.onerror = (error) => {
            console.error("WebSocket Error:", error);
            // readyState might already be CLOSING or CLOSED, update just in case
            readyState = socket.readyState;
        };

        // Cleanup on component unmount
        return () => {
            if (socket && socket.readyState < WebSocket.CLOSING) {
                console.log("Closing WebSocket connection on unmount");
                socket.close();
            }
            ws = null;
            readyState = null;
        };
    });
</script>

<div class="flex flex-col h-screen bg-gray-50 p-4">
    <div class="mb-4 p-2 bg-white rounded shadow">
        <p class="text-sm font-medium text-gray-500">CONNECTION:</p>
        <p class:text-green-600={readyState === 1} class:text-red-600={readyState !== 1} class="font-semibold">
            {readyStateString}
        </p>
    </div>

    <h1 class="text-2xl font-bold mb-4 text-gray-800">Simple Chat</h1>

    <!-- Message Display Area -->
    <div
        bind:this={messageContainer}
        onscroll={handleScroll}
        class="flex-grow bg-white rounded shadow overflow-y-auto p-4 mb-4 border border-gray-200 scroll-smooth"
    >
        <ul class="space-y-2">
            {#each messages as msg, i (i)}
                <li class="p-2 rounded bg-blue-50 text-gray-700">
                    {msg}
                </li>
            {/each}
            {#if messages.length === 0}
                <li class="text-gray-400 italic">No messages yet...</li>
            {/if}
        </ul>
    </div>

    <!-- Input Area -->
    <div class="mt-auto flex items-center space-x-2">
        <input
            type="text"
            bind:value={message}
            placeholder="Type your message..."
            class="flex-grow p-2 border border-gray-300 rounded shadow-sm focus:ring-blue-500 focus:border-blue-500"
            onkeydown={(e) => { if (e.key === 'Enter') sendMessage(); }}
        />
        <button
            onclick={sendMessage}
            disabled={readyState !== 1 || !message.trim()}
            class="px-4 py-2 bg-blue-600 text-white font-semibold rounded shadow hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
            Send
        </button>
    </div>
</div>