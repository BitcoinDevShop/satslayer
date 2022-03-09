<script lang="ts">
    import { browser } from "$app/env";
    import { onMount } from "svelte";

    let SATSLAYER = browser ? new Audio("/SATSLAYER.mp3") : null;
    function playSound() {
        // e.preventDefault();
        SATSLAYER.play();
    }
    function pay(e) {
        e.preventDefault();
        SATSLAYER.muted = true;
        SATSLAYER.play().then(resetAudio);
        sendPayment()
            .then((res) => {
                console.log(res);
                // playSound();
            })
            .catch((e) => {
                console.error(e);
            });
        // setTimeout(() => {
        //     console.log("play");
        //     playSound();
        // }, 5000);
    }

    let invoice = "";
    let interval = 1;

    function sub(e) {
        SATSLAYER.muted = true;
        SATSLAYER.play().then(resetAudio);
        e.preventDefault();
        setInterval(() => {
            console.log("play");
            playSound();
        }, interval * 1000);
    }

    function resetAudio() {
        SATSLAYER.pause();
        SATSLAYER.currentTime = 0;
        SATSLAYER.muted = false;
    }

    const clientBaseUrl = "http://localhost:8080";
    const serverBaseUrl = "http://localhost:8083";

    async function getInvoice() {
        const response = await fetch(`${serverBaseUrl}/getinvoice`);

        console.debug(response);

        const data = response.text();
        return data;
    }
    async function sendPayment() {
        const payload = {
            pull_interval: 0, // seconds, 0 means a one time payment
            invoice: invoice,
        };

        console.debug(payload);

        // Default options are marked with *
        const response = await fetch(`${clientBaseUrl}/sendpayment`, {
            method: "POST", // *GET, POST, PUT, DELETE, etc.
            // headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload), // body data type must match "Content-Type" header
        });

        const data = response.text();

        // const isJson = response.headers
        //     .get("content-type")
        //     ?.includes("application/json");
        // const data = isJson && (await response.json());

        // check for error response
        if (!response.ok) {
            // get error message from body or default to response status
            const error = (data && data.message) || response.status;
            return Promise.reject(error);
        }

        return data;
    }

    let socket;
    let slaycount = 0;
    // let websocketMessages = ["hello", "hello"];

    onMount(async () => {
        invoice = await getInvoice();
        socket = new WebSocket("ws://localhost:8083/subscribetx");
        socket.addEventListener("open", () => {
            console.log("Opened");
        });
        socket.addEventListener("message", () => {
            console.log("Opened");
            slaycount = slaycount + 1;
            playSound();

            // websocketMessages = [...websocketMessages, "hello"];
        });
    });
</script>

<main
    class="container h-screen mx-auto max-w-prose flex flex-col items-center justify-center pb-4"
>
    <h1 class="text-4xl font-black">SAT SLAYER</h1>
    <div class="h-4" />
    <pre class="break-all" />
    <p class="font-bold self-start ml-4">Slay count</p>
    <pre class="break-all text-8xl">{slaycount}</pre>

    <div class="h-4" />
    <p class="font-bold self-start ml-4">AMP invoice</p>
    {#if invoice === ""}
        <pre class="break-all">loading...</pre>
    {:else}
        <!-- <pre class="break-all">amp invoice:</pre> -->
        <pre class="break-all">{invoice}</pre>
    {/if}
    <div class="w-full flex">
        <div class="button">
            <button on:click={pay} class="bg-[#9DD1F1]"
                >{slaycount > 0 ? "Pay Again" : "Pay Once"}</button
            >
        </div>
        <div class="flex flex-col mt-8">
            <span>or</span>
            <div />
        </div>

        <div class="button">
            <button on:click={sub} class="bg-[#EFF7CF]">Subscribe</button>
            <label for="interval"> Interval (seconds)</label>
            <input
                id="interval"
                bind:value={interval}
                type="number"
                placeholder="interval (seconds)"
            />
        </div>
    </div>
</main>

<style lang="postcss">
    code {
    }

    pre {
        padding: 1rem;
        white-space: pre-wrap;
    }
    .button {
        width: 25ch;
        flex: 1;
        display: flex;
        flex-direction: column;
    }

    button {
        @apply m-4 text-black rounded shadow transition;
        @apply text-xl font-bold p-4;
    }

    button:hover,
    input:hover {
        @apply drop-shadow-[0px_0px_5px_rgba(192,87,70,0.75)];
    }

    button:active {
        opacity: 0.75;
    }

    input {
        @apply text-xl text-black p-2 m-4 rounded transition;
    }

    label {
        @apply ml-4 font-bold;
    }
</style>
