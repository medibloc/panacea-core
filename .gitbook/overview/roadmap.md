# Roadmap

![](images/roadmap.png)

Until 2021, MediBloc focused on upgrading Panacea to be more secure and interconnectable with other chains in Cosmos ecosystem. Inter Blockchain Communication (IBC) was enabled after the chain was upgraded with Cosmos SDK Stargate, and MED became tradable on one of the biggest decentralized exchange, Osmosis. On top of this, many professional validators and delegators have joined Panacea to enhance the security of the chain.

Now, MediBloc would like to shift its focus to boosting the potential of Panacea ecosystem and we believe that fruitful and meaningful data exchange will achieve that goal. Hence, MediBloc Panacea Team has been designing the Data Exchange Protocol based on Panacea. As you know, MediBloc believes that healthcare data should be owned by patients and those patients can take high-quality services by providing their healthcare data securely. Based on this philosophy, Data Exchange Protocol has been designed for anyone who holds data (even the sensitive data) to provide/sell their data to someone who is really willing to pay for it.

Nowadays, many businesses and technologies are data-driven. Many companies are already familiar with handling large dataset and deriving new values by analyzing sets of data. But, secure data exchange is still the one of the hardest area for data-driven industries. Data requesters want well-refined data or fine-grained raw data for successful data analysis. But, data owners (individuals) don’t want their privacy exposed and abused. Additionally, Web3 users are already aware that proper rewards should be guaranteed for their data and actions transparently on Web3. Traditional systems in Web2 have solved this issue in various ways, but MediBloc believes that we all can build more transparent and reliable systems for secure data exchange in Web3 ecosystem. 

Our data exchange protocol has the concept of data Pool, so that anyone can specify the type and the quantity of the data they want. Also, they can specify how much cryptocurrency they are willing to pay for the data. All of these data pools are recorded in Panacea and everyone who wants to sell their data can see all data pools. Data sellers can choose data pools by checking how many parts of their data to be shared to data buyers. Then, they sign the consents for data exchange. Verified off-chain data validators validate whether data provided by data sellers conforms to criteria that data pool creator has specified. If all the requirements are met, the data is provided to data buyers via secure connections and the promised amount of cryptocurrency is transferred to data sellers. In this entire protocol, data is not recorded on any blockchain such as Panacea. All data transmissions are performed off-chain and Panacea guarantees all agreements for data exchanges and transparent payments.

This data exchange protocol is being developed to be as general as possible, so that not only the healthcare data but also all the other types of data can be handled by the protocol. Since Panacea and data exchange protocol is publicly opened, any service providers can build their own services on the top of the data exchange protocol, so that their users can exchange their data securely and get proper rewards. As the first use case, MediBloc is going to build a healthcare data marketplace service based on this protocol.
Well, it sounds like the protocol should work well, right? However, there are so many issues that we have to resolve. For privacy and security, data sellers should be able to expose only a small part of their data that is really desired by data buyers. Data transmission must be secure, so that anyone cannot steal data. In order to guarantee the right of data buyers, all criteria that data buyers specified has to be validated clearly before the payment is finalized. In addition, the ecosystem should be attractive enough for many data sellers and buyers to join. 

---

In order to resolve these challenges, the team is developing this data market protocol with several latest technologies. 

### Data quality validation in Secure Enclaves

Data provided by sellers should be validated according to criteria that buyers specified. This validation process can be performed on Panacea because Panacea is already a Byzantine fault-tolerant consensus system. However, we should not put sensitive data on any public blockchain even if the data is encrypted. So, we introduced a new role: Verified Off-chain Data Validators with Secure Enclaves. Secure enclave technologies such as Intel SGX empowers data validators to validate data without exposing the content of data outside of the enclave. It means that operators of the data validator do their job without looking into the content of data. 

### W3C Verifiable Credentials and Zero-knowledge Proof

The protocol should support various ways for data buyers to specify what types of data they want to buy. We believe that W3C Verifiable Credential may be one of the solutions. Based on this standard, selective disclosure can also be achieved by zero-knowledge proof. 

In addition to these technologies above, the data exchange protocol is going to provide rich feature sets for service providers who build services on the top of the data exchange protocol in various legal countries and industries. It will be helpful for them to build their service in compliance with their law for handling personal data. For example, the protocol is going to provide features for making consents to data sharing/exchange signed cryptographically.

### NFT and DAO

The protocol is going to adopt the power of NFT. The data pool access vouchers are going to be minted as NFTs, so that data buyers and investors can join easily. This will also increase the financial liquidity of a given pool. In addition, the protocol will be designed to be governed by DAO. Any DAO participants can improve the protocol by changing parameters and developing new features.

---

There will be more details that we have to solve, and we know that all of them cannot be achieved in one step. Hence, we will complete this big task step by step. In 2022, MediBloc will release the v0 of data exchange protocol as a proof of concept that includes only essential features. Also, a data marketplace web service will be introduced as a simple example service based on the protocol. Based on this proof of concepts, the data exchange protocol will be improved as v1 from 2023 with enhanced security and interoperability. MediBloc has already opened all source codes and progresses publicly on GitHub. We encourage anyone to join the project and share your insights. 
We are so excited and thrilled to share our vision to achieve our goal to become the world’s best patient centric health data platform. Thank you for your continued support! 
