# Distributed heavy models using LLamaCPP

Is it possible for llamacpp to use multiple computers in single network to run some heavier model which single computer would not be able to run alone?

## Solution

Yes, it is absolutely possible to use `llama.cpp` across multiple computers on a single network to run models too large for any single machine. This is a supported feature called **distributed inference**, and it works surprisingly well.

The most straightforward way to achieve this is using `llama.cpp`'s built-in **RPC (Remote Procedure Call) backend**.

### 🤔 How Does It Work?
The system uses a simple "master-worker" architecture that is surprisingly easy to set up:

*   **The Master Node**: This is the computer you interact with. It loads the model (or a portion of it), runs the main `llama.cpp` program, and coordinates the workload.
*   **The Worker Nodes**: These are the other computers on your network. They run a lightweight `rpc-server` program and wait for instructions. The master node sends parts of the model (like layers of the neural network) to these workers, which perform the computation and send the results back.

In practice, `llama.cpp` distributes the model by **splitting it across layers**. Each worker machine is responsible for computing a specific set of layers. This approach minimizes the amount of data that needs to be transferred between machines at each step, making it feasible even over a standard network connection.

### 🚀 How to Set It Up
The setup process involves building `llama.cpp` with RPC support and then running a server on your worker machines. Here is the essential workflow, based on community guides:

**1. Build `llama.cpp` with RPC on all machines:**
   First, you need to compile `llama.cpp` with the RPC option enabled. This is done using `cmake`:
   ```bash
   cmake -B build -DGGML_RPC=ON
   cmake --build build --config Release -j
   ```
   If your machines have GPUs (like NVIDIA or AMD), you can also enable CUDA or ROCm support for a significant speed boost.

**2. Start the RPC Server on your Worker Machines:**
   On each computer you want to use as a worker, navigate to your `llama.cpp` build folder and run the `rpc-server` program. Make a note of its IP address on your local network (e.g., `192.168.1.10`).
   ```bash
   ./build/bin/rpc-server --host 0.0.0.0 --port 50052
   ```
   The `--host 0.0.0.0` flag makes the server accessible to other machines on your network.

**3. Run the Master Node:**
   On your primary computer, you can now run the inference command. The key is the `--rpc` flag, where you list the IP addresses and ports of all your worker servers.
   ```bash
   ./build/bin/llama-cli \
       -m /path/to/your/model.gguf \
       -p "Your prompt here" \
       --rpc 192.168.1.10:50052,192.168.1.11:50052 \
       -ngl 99
   ```
   The `-ngl 99` flag tells `llama.cpp` to offload as many layers as possible to the GPUs on the workers, if available.

### ⚡️ Performance and Practical Considerations
*   **Best for Large Models**: Distributed inference shines when a model is too large to fit on your most powerful machine. While there is a network overhead, it's a small price to pay for being able to run a massive model at all.
*   **Network Speed Matters**: A fast local network (Gigabit Ethernet or higher) is recommended. The performance bottleneck will often be the network connection, especially for smaller models where the overhead is more noticeable.
*   **Security**: The RPC server is not secure and should **never be exposed to the open internet**. Only run it within your trusted local network, as it has no built-in authentication or encryption.

### 💎 Alternative Methods
While the RPC backend is the most common and user-friendly method, it's worth noting that `llama.cpp` has also experimented with **MPI (Message Passing Interface)** in the past. However, the RPC method has become the standard, actively maintained approach for distributed inference.

To sum it up, `llama.cpp`'s RPC feature provides a practical and effective way to pool the memory of multiple computers on your home network to run very large language models. It is an advanced but well-documented feature for enthusiasts looking to push the boundaries of local AI.

