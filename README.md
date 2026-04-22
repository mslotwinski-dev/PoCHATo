<div align="center">
  <img src="https://minecraft.wiki/images/Potato_JE3_BE2.png" alt="poCHATo Logo" width="150"/>

  # poCHATo
  
  **A secure, real-time, end-to-end encrypted client-server chat application written in Go.**

  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
  [![Gin Framework](https://img.shields.io/badge/Gin-Web_Framework-0088CC?style=flat)](https://gin-gonic.com/)
  [![WebSockets](https://img.shields.io/badge/WebSockets-Real_Time-blue?style=flat)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

## 📖 Overview

**poCHATo** is a comprehensive computer networks project designed to provide secure, real-time messaging between users. Built entirely in Go, it features a robust client-server architecture utilizing a RESTful API for standard operations and WebSockets (TCP/IP) for low-latency, bi-directional communication. 

Security and privacy are at the core of poCHATo. Direct communication is restricted to mutually accepted friends, and every single message is secured using **End-to-End Encryption (E2EE)**. 

## ✨ Key Features

* **🔒 End-to-End Encryption (E2EE):** Total privacy. Public and private key pairs are generated when users become friends. Messages are fully encrypted on the sender's device and decrypted only on the recipient's device. The server never reads the plaintext messages.
* **🤝 Friend System:** Users must establish a connection before chatting seamlessly. You can easily fetch your friend list and manage your connections.
* **⚡ Real-Time WebSockets:** Powered by TCP/IP and WebSockets for instant message delivery without continuous HTTP polling.
* **💬 Rich Chat Interactions:** (Heart messages (❤️)).
    * Block malicious or unwanted users.
* **🔐 Secure Authentication:** Robust user registration and login system with securely hashed passwords.

---

## 🏗️ Architecture & Tech Stack

poCHATo consists of a backend server and a client application, strictly separating concerns:

### Backend (Server)
* **Language:** Go (Golang)
* **API Framework:** [Gin](https://gin-gonic.com/) for fast RESTful routing (Login, Registration, Friend Management).
* **Real-time Protocol:** Gorilla WebSockets over TCP/IP for the active chat channels.
* **Database:** SQLite for storing users, hashed credentials, and encrypted message blobs.

### Frontend (Client)
* **Language:** Go (CLI or GUI depending on implementation)
* **Cryptography:** Standard Go `crypto/rsa` and `crypto/aes` for generating keys and handling E2EE.

---

## 🔐 How E2EE Works in poCHATo

To ensure the server cannot read messages, poCHATo implements a strict cryptographic flow:

1.  **Friend Request Accepted:** When User A and User B become friends, their clients generate standard RSA Public/Private key pairs.
2.  **Key Exchange:** Public keys are exchanged through the server and stored locally on the clients. Private keys *never* leave the user's local device.
3.  **Sending a Message:** When User A messages User B, User A's client encrypts the message payload using User B's Public Key.
4.  **Storage & Delivery:** The server receives and stores the ciphertext, delivering it via WebSockets to User B.
5.  **Decryption:** User B's client receives the ciphertext and decrypts it using User B's Private Key, revealing the original message.

---

## 🚀 Getting Started

### Prerequisites
* [Go](https://go.dev/dl/) 1.21 or higher installed.
* A running instance of your chosen Database (SQLite).

### Installation

1. Clone the repository:
   ```bash
   git clone [https://github.com/yourusername/poCHATo.git](https://github.com/yourusername/poCHATo.git)
   cd poCHATo
